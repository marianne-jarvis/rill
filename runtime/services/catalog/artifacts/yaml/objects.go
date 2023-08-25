package yaml

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/jinzhu/copier"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"google.golang.org/protobuf/types/known/structpb"

	// Load IANA time zone data
	_ "time/tzdata"
)

/**
 * This file contains the mapping from CatalogObject to Yaml files
 */
type Source struct {
	Type                  string
	Path                  string         `yaml:"path,omitempty"`
	CsvDelimiter          string         `yaml:"csv.delimiter,omitempty" mapstructure:"csv.delimiter,omitempty"`
	URI                   string         `yaml:"uri,omitempty"`
	Region                string         `yaml:"region,omitempty" mapstructure:"region,omitempty"`
	S3Endpoint            string         `yaml:"endpoint,omitempty" mapstructure:"endpoint,omitempty"`
	GlobMaxTotalSize      int64          `yaml:"glob.max_total_size,omitempty" mapstructure:"glob.max_total_size,omitempty"`
	GlobMaxObjectsMatched int            `yaml:"glob.max_objects_matched,omitempty" mapstructure:"glob.max_objects_matched,omitempty"`
	GlobMaxObjectsListed  int64          `yaml:"glob.max_objects_listed,omitempty" mapstructure:"glob.max_objects_listed,omitempty"`
	GlobPageSize          int            `yaml:"glob.page_size,omitempty" mapstructure:"glob.page_size,omitempty"`
	HivePartition         *bool          `yaml:"hive_partitioning,omitempty" mapstructure:"hive_partitioning,omitempty"`
	Timeout               int32          `yaml:"timeout,omitempty"`
	ExtractPolicy         *ExtractPolicy `yaml:"extract,omitempty"`
	Format                string         `yaml:"format,omitempty" mapstructure:"format,omitempty"`
	DuckDBProps           map[string]any `yaml:"duckdb,omitempty" mapstructure:"duckdb,omitempty"`
	Headers               map[string]any `yaml:"headers,omitempty" mapstructure:"headers,omitempty"`
	AllowSchemaRelaxation *bool          `yaml:"ingest.allow_schema_relaxation,omitempty" mapstructure:"allow_schema_relaxation,omitempty"`
	SQL                   string         `yaml:"sql,omitempty" mapstructure:"sql,omitempty"`
	DB                    string         `yaml:"db,omitempty" mapstructure:"db,omitempty"`
	ProjectID             string         `yaml:"project_id,omitempty" mapstructure:"project_id,omitempty"`
}

type ExtractPolicy struct {
	Row  *ExtractConfig `yaml:"rows,omitempty" mapstructure:"rows,omitempty"`
	File *ExtractConfig `yaml:"files,omitempty" mapstructure:"files,omitempty"`
}

type ExtractConfig struct {
	Strategy string `yaml:"strategy,omitempty" mapstructure:"strategy,omitempty"`
	Size     string `yaml:"size,omitempty" mapstructure:"size,omitempty"`
}

type MetricsView struct {
	Label              string `yaml:"title"`
	DisplayName        string `yaml:"display_name,omitempty"` // for backwards compatibility
	Description        string
	Model              string
	TimeDimension      string   `yaml:"timeseries"`
	SmallestTimeGrain  string   `yaml:"smallest_time_grain"`
	DefaultTimeRange   string   `yaml:"default_time_range"`
	AvailableTimeZones []string `yaml:"available_time_zones,omitempty"`
	Dimensions         []*Dimension
	Measures           []*Measure
	Policy             *Policy `yaml:"policy,omitempty"`
}

type Policy struct {
	HasAccess string               `yaml:"has_access,omitempty"`
	Filter    string               `yaml:"filter,omitempty"`
	Include   []*ConditionalColumn `yaml:"include,omitempty"`
	Exclude   []*ConditionalColumn `yaml:"exclude,omitempty"`
}

type ConditionalColumn struct {
	Name      string
	Condition string `yaml:"if"`
}

type Measure struct {
	Label               string
	Name                string
	Expression          string
	Description         string
	Format              string `yaml:"format_preset"`
	Ignore              bool   `yaml:"ignore,omitempty"`
	ValidPercentOfTotal bool   `yaml:"valid_percent_of_total,omitempty"`
}

type Dimension struct {
	Name        string
	Label       string
	Property    string `yaml:"property,omitempty"`
	Column      string
	Description string
	Ignore      bool `yaml:"ignore,omitempty"`
}

func toSourceArtifact(catalog *drivers.CatalogEntry) (*Source, error) {
	source := &Source{
		Type: catalog.GetSource().Connector,
	}

	props := catalog.GetSource().Properties.AsMap()

	err := mapstructure.Decode(props, source)
	if err != nil {
		return nil, err
	}

	if source.Path != "" && catalog.GetSource().Connector != "local_file" {
		source.URI = source.Path
		source.Path = ""
	}

	extract, err := toExtractArtifact(catalog.GetSource().GetPolicy())
	if err != nil {
		return nil, err
	}

	source.ExtractPolicy = extract
	return source, nil
}

func toExtractArtifact(extract *runtimev1.Source_ExtractPolicy) (*ExtractPolicy, error) {
	if extract == nil {
		return nil, nil
	}

	sourceExtract := &ExtractPolicy{}
	// set file
	if extract.FilesStrategy != runtimev1.Source_ExtractPolicy_STRATEGY_UNSPECIFIED {
		sourceExtract.File = &ExtractConfig{}
		sourceExtract.File.Strategy = extract.FilesStrategy.String()
		sourceExtract.File.Size = fmt.Sprintf("%v", extract.FilesLimit)
	}

	// set row
	if extract.RowsStrategy != runtimev1.Source_ExtractPolicy_STRATEGY_UNSPECIFIED {
		sourceExtract.Row = &ExtractConfig{}
		sourceExtract.Row.Strategy = extract.RowsStrategy.String()
		bytes := datasize.ByteSize(extract.RowsLimitBytes)
		sourceExtract.Row.Size = bytes.HumanReadable()
	}

	return sourceExtract, nil
}

func toMetricsViewArtifact(catalog *drivers.CatalogEntry) (*MetricsView, error) {
	metricsArtifact := &MetricsView{}
	err := copier.Copy(metricsArtifact, catalog.Object)
	metricsArtifact.SmallestTimeGrain = getTimeGrainString(catalog.GetMetricsView().SmallestTimeGrain)
	metricsArtifact.DefaultTimeRange = catalog.GetMetricsView().DefaultTimeRange
	if err != nil {
		return nil, err
	}

	return metricsArtifact, nil
}

func fromSourceArtifact(source *Source, path string) (*drivers.CatalogEntry, error) {
	props := map[string]interface{}{}
	if source.Type == "local_file" {
		props["path"] = source.Path
	} else if source.URI != "" {
		props["path"] = source.URI
	}
	if source.Region != "" {
		props["region"] = source.Region
	}

	if source.DuckDBProps != nil {
		props["duckdb"] = source.DuckDBProps
	}

	if source.CsvDelimiter != "" {
		// backward compatibility
		if _, defined := props["duckdb"]; !defined {
			props["duckdb"] = map[string]any{}
		}
		props["duckdb"].(map[string]any)["delim"] = fmt.Sprintf("'%v'", source.CsvDelimiter)
	}

	if source.HivePartition != nil {
		// backward compatibility
		if _, defined := props["duckdb"]; !defined {
			props["duckdb"] = map[string]any{}
		}
		props["duckdb"].(map[string]any)["hive_partitioning"] = *source.HivePartition
	}

	if source.GlobMaxTotalSize != 0 {
		props["glob.max_total_size"] = source.GlobMaxTotalSize
	}

	if source.GlobMaxObjectsMatched != 0 {
		props["glob.max_objects_matched"] = source.GlobMaxObjectsMatched
	}

	if source.GlobMaxObjectsListed != 0 {
		props["glob.max_objects_listed"] = source.GlobMaxObjectsListed
	}

	if source.GlobPageSize != 0 {
		props["glob.page_size"] = source.GlobPageSize
	}

	if source.S3Endpoint != "" {
		props["endpoint"] = source.S3Endpoint
	}

	if source.Format != "" {
		props["format"] = source.Format
	}

	if source.Headers != nil {
		props["headers"] = source.Headers
	}

	if source.AllowSchemaRelaxation != nil {
		props["allow_schema_relaxation"] = *source.AllowSchemaRelaxation
	}

	if source.SQL != "" {
		props["sql"] = source.SQL
	}

	if source.DB != "" {
		props["db"] = source.DB
	}

	if source.ProjectID != "" {
		props["project_id"] = source.ProjectID
	}

	propsPB, err := structpb.NewStruct(props)
	if err != nil {
		return nil, err
	}

	extract, err := fromExtractArtifact(source.ExtractPolicy)
	if err != nil {
		return nil, err
	}

	name := fileutil.Stem(path)
	return &drivers.CatalogEntry{
		Name: name,
		Type: drivers.ObjectTypeSource,
		Path: path,
		Object: &runtimev1.Source{
			Name:           name,
			Connector:      source.Type,
			Properties:     propsPB,
			Policy:         extract,
			TimeoutSeconds: source.Timeout,
		},
	}, nil
}

func fromExtractArtifact(policy *ExtractPolicy) (*runtimev1.Source_ExtractPolicy, error) {
	if policy == nil {
		return nil, nil
	}

	extractPolicy := &runtimev1.Source_ExtractPolicy{}

	// parse file
	if policy.File != nil {
		// parse strategy
		strategy, err := parseStrategy(policy.File.Strategy)
		if err != nil {
			return nil, err
		}

		extractPolicy.FilesStrategy = strategy

		// parse size
		size, err := strconv.ParseUint(policy.File.Size, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid size, parse failed with error %w", err)
		}
		if size <= 0 {
			return nil, fmt.Errorf("invalid size %q", size)
		}

		extractPolicy.FilesLimit = size
	}

	// parse rows
	if policy.Row != nil {
		// parse strategy
		strategy, err := parseStrategy(policy.Row.Strategy)
		if err != nil {
			return nil, err
		}

		extractPolicy.RowsStrategy = strategy

		// parse size
		// todo :: add support for number of rows
		size, err := getBytes(policy.Row.Size)
		if err != nil {
			return nil, fmt.Errorf("invalid size, parse failed with error %w", err)
		}
		if size <= 0 {
			return nil, fmt.Errorf("invalid size %q", size)
		}

		extractPolicy.RowsLimitBytes = size
	}
	return extractPolicy, nil
}

func parseStrategy(s string) (runtimev1.Source_ExtractPolicy_Strategy, error) {
	switch strings.ToLower(s) {
	case "tail":
		return runtimev1.Source_ExtractPolicy_STRATEGY_TAIL, nil
	case "head":
		return runtimev1.Source_ExtractPolicy_STRATEGY_HEAD, nil
	default:
		return runtimev1.Source_ExtractPolicy_STRATEGY_UNSPECIFIED, fmt.Errorf("invalid extract strategy %q", s)
	}
}

func getBytes(size string) (uint64, error) {
	var s datasize.ByteSize
	if err := s.UnmarshalText([]byte(size)); err != nil {
		return 0, err
	}

	return s.Bytes(), nil
}

func fromMetricsViewArtifact(metrics *MetricsView, path string) (*drivers.CatalogEntry, error) {
	if metrics.DisplayName != "" && metrics.Label == "" {
		// backwards compatibility
		metrics.Label = metrics.DisplayName
	}

	// remove ignored measures and dimensions
	var measures []*Measure
	for _, measure := range metrics.Measures {
		if measure.Ignore {
			continue
		}
		measures = append(measures, measure)
	}
	metrics.Measures = measures

	// create a map of dimensions to be used for policy validation
	dimensionsMap := map[string]bool{}

	var dimensions []*Dimension
	for _, dimension := range metrics.Dimensions {
		if dimension.Ignore {
			continue
		}
		if dimension.Property != "" && dimension.Column == "" {
			// backwards compatibility when we were using `property` instead of `column`
			dimension.Column = dimension.Property
		}
		dimensions = append(dimensions, dimension)
		dimensionsMap[dimension.Column] = true
	}
	metrics.Dimensions = dimensions

	apiMetrics := &runtimev1.MetricsView{}

	// validate correctness of default time range
	if metrics.DefaultTimeRange != "" {
		_, err := duration.ParseISO8601(metrics.DefaultTimeRange)
		if err != nil {
			return nil, fmt.Errorf("invalid default_time_range: %w", err)
		}
		apiMetrics.DefaultTimeRange = metrics.DefaultTimeRange
	}

	// validate time zone locations
	for _, tz := range metrics.AvailableTimeZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return nil, err
		}
	}

	if metrics.Policy != nil {
		templateData := rillv1.TemplateData{User: map[string]interface{}{
			"name":   "dummy",
			"email":  "mock@example.org",
			"domain": "example.org",
			"groups": []interface{}{"all"},
			"admin":  false,
		}}

		if metrics.Policy.HasAccess != "" {
			hasAccess, err := rillv1.ResolveTemplate(metrics.Policy.HasAccess, templateData)
			if err != nil {
				return nil, fmt.Errorf(`invalid 'policy': 'has_access' templating is not valid: %w`, err)
			}
			_, err = rillv1.EvaluateBoolExpression(hasAccess)
			if err != nil {
				return nil, fmt.Errorf(`invalid 'policy': 'has_access' expression not valuating to a boolean: %w`, err)
			}
		}

		if metrics.Policy.Filter != "" {
			_, err := rillv1.ResolveTemplate(metrics.Policy.Filter, templateData)
			if err != nil {
				return nil, fmt.Errorf(`invalid 'policy': 'filter' templating is not valid: %w`, err)
			}
		}

		if len(metrics.Policy.Include) > 0 && len(metrics.Policy.Exclude) > 0 {
			return nil, errors.New("invalid 'policy': only one of 'include' and 'exclude' can be specified")
		}

		err := validatedPolicyFieldList(metrics.Policy.Include, dimensionsMap, "include", templateData)
		if err != nil {
			return nil, err
		}

		err = validatedPolicyFieldList(metrics.Policy.Exclude, dimensionsMap, "exclude", templateData)
		if err != nil {
			return nil, err
		}
	}

	err := copier.Copy(apiMetrics, metrics)
	if err != nil {
		return nil, err
	}

	// this is needed since measure names are not given by the user
	for i, measure := range apiMetrics.Measures {
		if measure.Name == "" {
			measure.Name = fmt.Sprintf("measure_%d", i)
		}
	}

	// backwards compatibility where name was used as property
	for i, dimension := range apiMetrics.Dimensions {
		if dimension.Name == "" {
			if dimension.Column == "" {
				// if there is no name and property add dimension_<index> as name
				dimension.Name = fmt.Sprintf("dimension_%d", i)
			} else {
				// else use property as name
				dimension.Name = dimension.Column
			}
		}
	}

	timeGrainEnum, err := getTimeGrainEnum(metrics.SmallestTimeGrain)
	if err != nil {
		return nil, err
	}
	apiMetrics.SmallestTimeGrain = timeGrainEnum

	name := fileutil.Stem(path)
	apiMetrics.Name = name
	return &drivers.CatalogEntry{
		Name:   name,
		Type:   drivers.ObjectTypeMetricsView,
		Path:   path,
		Object: apiMetrics,
	}, nil
}

func validatedPolicyFieldList(fieldConditions []*ConditionalColumn, dimensions map[string]bool, property string, templateData rillv1.TemplateData) error {
	if len(fieldConditions) > 0 {
		for _, field := range fieldConditions {
			if field.Name == "" || field.Condition == "" {
				return fmt.Errorf("invalid 'policy': '%s' fields must have a valid 'name' and 'if' condition", property)
			}
			// check the name is a valid dimension that exists in the metrics view
			if !dimensions[field.Name] {
				return fmt.Errorf("invalid 'policy': '%s' property %q does not exists in dimensions list", property, field.Name)
			}
			cond, err := rillv1.ResolveTemplate(field.Condition, templateData)
			if err != nil {
				return fmt.Errorf(`invalid 'policy': 'if' condition templating for field %q is not valid: %w`, field.Name, err)
			}
			_, err = rillv1.EvaluateBoolExpression(cond)
			if err != nil {
				return fmt.Errorf(`invalid 'policy': 'if' condition for field %q not valuating to a boolean: %w`, field.Name, err)
			}
		}
	}
	return nil
}

// Get TimeGrain enum from string
func getTimeGrainEnum(timeGrain string) (runtimev1.TimeGrain, error) {
	switch strings.ToLower(timeGrain) {
	case "":
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, nil
	case "millisecond":
		return runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, nil
	case "second":
		return runtimev1.TimeGrain_TIME_GRAIN_SECOND, nil
	case "minute":
		return runtimev1.TimeGrain_TIME_GRAIN_MINUTE, nil
	case "hour":
		return runtimev1.TimeGrain_TIME_GRAIN_HOUR, nil
	case "day":
		return runtimev1.TimeGrain_TIME_GRAIN_DAY, nil
	case "week":
		return runtimev1.TimeGrain_TIME_GRAIN_WEEK, nil
	case "month":
		return runtimev1.TimeGrain_TIME_GRAIN_MONTH, nil
	case "quarter":
		return runtimev1.TimeGrain_TIME_GRAIN_QUARTER, nil
	case "year":
		return runtimev1.TimeGrain_TIME_GRAIN_YEAR, nil
	default:
		return runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED, fmt.Errorf("invalid time grain: %s", timeGrain)
	}
}

// Get TimeGrain string from enum
func getTimeGrainString(timeGrain runtimev1.TimeGrain) string {
	switch timeGrain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "millisecond"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "second"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "minute"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "hour"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "day"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "week"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "month"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "quarter"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "year"
	default:
		return ""
	}
}
