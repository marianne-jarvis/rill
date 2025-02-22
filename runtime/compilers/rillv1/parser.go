package rillv1

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/exp/slices"
)

// Built-in parser limits
var (
	maxFiles    = 10000
	maxFileSize = 8192 // 8kb
)

// ErrInvalidProject indicates a project without a rill.yaml file
var ErrInvalidProject = errors.New("parser: not a valid project (rill.yaml not found)")

// Resource parsed from code files.
// One file may output multiple resources and multiple files may contribute config to one resource.
type Resource struct {
	// Metadata
	Name    ResourceName
	Paths   []string
	Refs    []ResourceName // Derived from rawRefs after parsing (can't contain ResourceKindUnspecified)
	rawRefs []ResourceName // Populated during parsing (may contain ResourceKindUnspecified)

	// Only one of these will be non-nil
	SourceSpec      *runtimev1.SourceSpec
	ModelSpec       *runtimev1.ModelSpec
	MetricsViewSpec *runtimev1.MetricsViewSpec
	MigrationSpec   *runtimev1.MigrationSpec
}

// ResourceName is a unique identifier for a resource
type ResourceName struct {
	Kind ResourceKind
	Name string
}

func (n ResourceName) String() string {
	return fmt.Sprintf("%s/%s", n.Kind, n.Name)
}

func (n ResourceName) Normalized() ResourceName {
	return ResourceName{
		Kind: n.Kind,
		Name: strings.ToLower(n.Name),
	}
}

// ResourceKind identifies a resource type supported by the parser
type ResourceKind int

const (
	ResourceKindUnspecified ResourceKind = iota
	ResourceKindSource
	ResourceKindModel
	ResourceKindMetricsView
	ResourceKindMigration
)

// ParseResourceKind maps a string to a ResourceKind.
// Note: The empty string is considered a valid kind (unspecified).
func ParseResourceKind(kind string) (ResourceKind, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "":
		return ResourceKindUnspecified, nil
	case "source":
		return ResourceKindSource, nil
	case "model":
		return ResourceKindModel, nil
	case "metricsview", "metrics_view", "dashboard":
		return ResourceKindMetricsView, nil
	case "migration":
		return ResourceKindMigration, nil
	default:
		return ResourceKindUnspecified, fmt.Errorf("invalid resource kind %q", kind)
	}
}

func (k ResourceKind) String() string {
	switch k {
	case ResourceKindUnspecified:
		return ""
	case ResourceKindSource:
		return "Source"
	case ResourceKindModel:
		return "Model"
	case ResourceKindMetricsView:
		return "MetricsView"
	case ResourceKindMigration:
		return "Migration"
	default:
		panic(fmt.Sprintf("unexpected resource kind: %d", k))
	}
}

// Diff shows changes to Parser.Resources following an incremental reparse.
type Diff struct {
	Added            []ResourceName
	Modified         []ResourceName
	ModifiedRillYAML bool
	Deleted          []ResourceName
}

// Parser parses a Rill project directory into a set of resources.
// After the initial parse, the parser can be used to incrementally reparse a subset of files.
// Parser is not concurrency safe.
type Parser struct {
	// Options
	Repo             drivers.RepoStore
	InstanceID       string
	DuckDBConnectors []string

	// Output
	RillYAML  *RillYAML
	Resources map[ResourceName]*Resource
	Errors    []*runtimev1.ParseError

	// Internal state
	resourcesForPath           map[string][]*Resource // Reverse index of Resource.Paths
	resourcesForUnspecifiedRef map[string][]*Resource // Reverse index of Resource.rawRefs where kind=ResourceKindUnspecified
	insertedResources          []*Resource
	updatedResources           []*Resource
}

// ParseRillYAML parses only the project's rill.yaml (or rill.yml) file.
func ParseRillYAML(ctx context.Context, repo drivers.RepoStore, instanceID string) (*RillYAML, error) {
	paths, err := repo.ListRecursive(ctx, "rill.{yaml,yml}")
	if err != nil {
		return nil, fmt.Errorf("could not list project files: %w", err)
	}

	p := Parser{Repo: repo, InstanceID: instanceID}
	err = p.parsePaths(ctx, paths)
	if err != nil {
		return nil, err
	}

	return p.RillYAML, nil
}

// Parse creates a new parser and parses the entire project.
//
// Note on SQL parsing: For DuckDB SQL specifically, the parser can use a SQL parser to extract refs and annotations (instead of relying on templating or YAML).
// To enable SQL parsing for a connector, pass it in duckDBConnectors. If DuckDB SQL parsing should be used on files where no connector is specified, put an empty string in duckDBConnectors.
func Parse(ctx context.Context, repo drivers.RepoStore, instanceID string, duckDBConnectors []string) (*Parser, error) {
	p := &Parser{
		Repo:                       repo,
		InstanceID:                 instanceID,
		DuckDBConnectors:           duckDBConnectors,
		Resources:                  make(map[ResourceName]*Resource),
		resourcesForPath:           make(map[string][]*Resource),
		resourcesForUnspecifiedRef: make(map[string][]*Resource),
	}

	paths, err := p.Repo.ListRecursive(ctx, "**/*.{sql,yaml,yml}")
	if err != nil {
		return nil, fmt.Errorf("could not list project files: %w", err)
	}

	err = p.parsePaths(ctx, paths)
	if err != nil {
		return nil, err
	}

	// Infer unspecified refs for all inserted resources
	for _, r := range p.insertedResources {
		p.inferUnspecifiedRefs(r)
	}

	return p, nil
}

// Reparse re-parses the indicated file paths, updating the Parser's state.
// If a previous call to Reparse has returned an error, the Parser may not be accessed or called again.
func (p *Parser) Reparse(ctx context.Context, paths []string) (*Diff, error) {
	// The logic here is slightly tricky because the relationship between files and resources can vary:
	//
	// - Case 1: one file created one resource
	// - Case 2: one file created multiple resources
	// - Case 3: multiple files contributed to one resource (for example, "model.sql" and "model.yaml")
	//
	// The high-level approach is: We'll delete all existing resources *related* to the changed paths and (re)parse them.
	// Then at the end, we build a diff that treats any resource that was both "deleted" and "added" as an "update".
	// (Renames are not supported in the parser. It needs to be handled by the caller, since parser state alone is insufficient to detect it.)
	//
	// Another wrinkle is that we need to re-infer unspecified refs for:
	// - any resources pointing to a changed resource, and
	// - any resources with previously unmatched unspecified refs that may match a new resource.

	// Reset insertedResources and updatedResources on reparse (used to construct Diff)
	p.insertedResources = nil
	p.updatedResources = nil

	// Phase 1: Clear existing state related to the paths.
	// Identify all paths directly passed and paths indirectly related through resourcesForPath and Resource.Paths.
	// And delete all resources and parse errors related to those paths.
	var parsePaths []string            // Paths we should pass to parsePaths
	var deletedResources []*Resource   // Resources deleted in Phase 1 (some may be added back in Phase 2)
	checkPaths := slices.Clone(paths)  // Paths we should visit in the loop
	seenPaths := make(map[string]bool) // Paths already visited by the loop
	for i := 0; i < len(checkPaths); i++ {
		// Don't check the same path twice
		path := normalizePath(checkPaths[i])
		if seenPaths[path] {
			continue
		}
		seenPaths[path] = true

		// Skip files that aren't SQL or YAML
		isSQL := strings.HasSuffix(path, ".sql")
		isYAML := strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
		if !isSQL && !isYAML {
			continue
		}

		// If a file exists at path, add it to the parse list
		_, err := p.Repo.Stat(ctx, path)
		if err == nil {
			parsePaths = append(parsePaths, path)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("unexpected file stat error: %w", err)
		}

		// Check if path is rill.yaml and clear it (so we can re-parse it)
		if path == "/rill.yaml" || path == "/rill.yml" {
			p.RillYAML = nil
		}

		// Since .sql and .yaml files provide context for each other, if one was modified, we need to reparse both.
		// For cases where a file was modified or deleted, the transitive check through resourcesForPath will already take of that.
		// But this ensures the check also happens for cases where a companion file was added.
		stem := pathStem(path)
		if isSQL {
			checkPaths = append(checkPaths, stem+".yaml", stem+".yml")
		} else if isYAML {
			checkPaths = append(checkPaths, stem+".sql")
		}

		// Remove all resources derived from this path, and add any related paths to the check list
		rs := slices.Clone(p.resourcesForPath[path]) // Use Clone because deleteResource mutates it
		for _, resource := range rs {
			p.deleteResource(resource)
			deletedResources = append(deletedResources, resource)

			// Make sure we-reparse all paths that contributed to the deleted resource.
			checkPaths = append(checkPaths, resource.Paths...)
		}

		// Remove all parse errors related to this path
		// (We can't mutate p.Errors while iterating over it, hence the nested loop here.)
		for {
			found := false
			for i, err := range p.Errors {
				if err.FilePath == path {
					found = true
					p.Errors = slices.Delete(p.Errors, i, i+1)
					break
				}
			}
			if !found {
				break
			}
		}
	}

	// Capture if rill.yaml will be updated
	modifiedRillYAML := p.RillYAML == nil

	// Phase 2: Parse (or reparse) the related paths, adding back resources
	err := p.parsePaths(ctx, parsePaths)
	if err != nil {
		return nil, err
	}

	// Infer or re-infer refs for...
	inferRefsSeen := make(map[ResourceName]bool)
	// ... all inserted resources
	for _, r := range p.insertedResources {
		inferRefsSeen[r.Name.Normalized()] = true
		p.inferUnspecifiedRefs(r)
	}
	// ... all updated resources
	for _, r := range p.updatedResources {
		inferRefsSeen[r.Name.Normalized()] = true
		p.inferUnspecifiedRefs(r)
	}
	// ... any unchanged resource that may have an unspecified ref to a deleted resource
	for _, r1 := range deletedResources {
		for _, r2 := range p.resourcesForUnspecifiedRef[strings.ToLower(r1.Name.Name)] {
			n := r2.Name.Normalized()
			if !inferRefsSeen[n] {
				inferRefsSeen[n] = true
				p.inferUnspecifiedRefs(r2)
				p.updatedResources = append(p.updatedResources, r2) // inferRefsSeen ensures it's not already in insertedResources or updatedResources
			}
		}
	}
	// ... any unchanged resource that might have an unspecified ref (previously unmatched) that now matches a newly inserted resource
	for _, r1 := range p.insertedResources {
		for _, r2 := range p.resourcesForUnspecifiedRef[strings.ToLower(r1.Name.Name)] {
			n := r2.Name.Normalized()
			if !inferRefsSeen[n] {
				inferRefsSeen[n] = true
				p.inferUnspecifiedRefs(r2)
				p.updatedResources = append(p.updatedResources, r2) // inferRefsSeen ensures it's not already in insertedResources or updatedResources
			}
		}
	}

	// Phase 3: Build the diff using p.insertedResources, p.updatedResources and deletedResources
	diff := &Diff{ModifiedRillYAML: modifiedRillYAML}
	for _, resource := range p.insertedResources {
		addedBack := false
		for _, deleted := range deletedResources {
			if resource.Name.Normalized() == deleted.Name.Normalized() {
				addedBack = true
				break
			}
		}
		if addedBack {
			diff.Modified = append(diff.Modified, resource.Name)
		} else {
			diff.Added = append(diff.Added, resource.Name)
		}
	}
	for _, resource := range p.updatedResources {
		diff.Modified = append(diff.Modified, resource.Name)
	}
	for _, deleted := range deletedResources {
		if p.Resources[deleted.Name.Normalized()] == nil {
			diff.Deleted = append(diff.Deleted, deleted.Name)
		}
	}

	return diff, nil
}

// parsePaths is the internal entrypoint for parsing a list of paths.
// It assumes that the caller has already checked that the paths exist.
// It also assumes that the caller has already removed any previous resources related to the paths,
// enabling parsePaths to upsert changes, enabling multiple files to provide data for one resource
// (like "my-model.sql" and "my-model.yaml").
func (p *Parser) parsePaths(ctx context.Context, paths []string) error {
	// Check limits
	if len(paths) > maxFiles {
		return fmt.Errorf("project exceeds file limit of %d", maxFiles)
	}

	// Sort paths such that we align files with the same name but different extensions next to each other.
	// Then iterate over the sorted paths, processing all paths with the same stem at once (stem = path without extension).
	slices.Sort(paths)
	for i := 0; i < len(paths); {
		// Handle rill.yaml separately (if parsing of rill.yaml fails, we exit early instead of adding a ParseError)
		path := paths[i]
		if path == "/rill.yaml" || path == "/rill.yml" {
			err := p.parseRillYAML(ctx, path)
			if err != nil {
				return err
			}
			i++
			continue
		}

		// Identify the range of paths with the same stem as paths[i]
		j := i + 1
		pathStemI := pathStem(paths[i])
		for j < len(paths) && pathStemI == pathStem(paths[j]) {
			j++
		}

		// Parse the paths with the same stem
		err := p.parseStemPaths(ctx, paths[i:j])
		if err != nil {
			return err
		}

		// Advance i to the next stem
		i = j
	}

	// If we didn't encounter rill.yaml, that's a breaking error
	if p.RillYAML == nil {
		return ErrInvalidProject
	}

	// As a special case, we need to check that there aren't any sources and models with the same name.
	// NOTE 1: We always attach the error to the model when there's a collision.
	// NOTE 2: Using a map since the two-way check (necessary for reparses) may match the same resource twice.
	modelsWithNameErrs := make(map[ResourceName]string)
	for _, r := range p.insertedResources {
		if r.Name.Kind == ResourceKindSource {
			n := ResourceName{Kind: ResourceKindModel, Name: r.Name.Name}.Normalized()
			if _, ok := p.Resources[n]; ok {
				modelsWithNameErrs[n.Normalized()] = r.Name.Name
			}
		} else if r.Name.Kind == ResourceKindModel {
			n := ResourceName{Kind: ResourceKindSource, Name: r.Name.Name}.Normalized()
			if r2, ok := p.Resources[n]; ok {
				modelsWithNameErrs[r.Name.Normalized()] = r2.Name.Name
			}
		}
	}
	for n, s := range modelsWithNameErrs {
		p.replaceResourceWithError(n, fmt.Errorf("model name collides with source %q", s))
	}

	return nil
}

// parseStem parses a set of paths with the same stem (path without extension).
func (p *Parser) parseStemPaths(ctx context.Context, paths []string) error {
	// Load YAML and SQL contents
	var yaml, yamlPath, sql, sqlPath string
	for _, path := range paths {
		// Load contents
		data, err := p.Repo.Get(ctx, path)
		if err != nil {
			if os.IsNotExist(err) {
				// This is a dirty parse where a file disappeared during parsing.
				// But due to the clear-and-rebuild behavior, we can safely continue parsing.
				continue
			}
			return err
		}

		// Check size
		if len(data) > maxFileSize {
			p.Errors = append(p.Errors, &runtimev1.ParseError{
				Message:  fmt.Sprintf("size %d bytes exceeds max size of %d bytes", len(data), maxFileSize),
				FilePath: path,
			})
			continue
		}

		// Assign to correct variable
		if strings.HasSuffix(path, ".sql") {
			sql = data
			sqlPath = path
			continue
		}
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			if yaml != "" {
				// Means there was both a .yaml and .yml file. We don't allow that!
				p.Errors = append(p.Errors, &runtimev1.ParseError{
					Message:  "skipping file because another YAML file has already been parsed for this path stem",
					FilePath: path,
				})
				continue
			}
			yaml = data
			yamlPath = path
			continue
		}
		// The unhandled case should never happen, just being defensive
	}

	// Parse the SQL/YAML file pair to a Node, then parse the Node to p.Resources.
	node, err := p.parseStem(ctx, paths, yamlPath, yaml, sqlPath, sql)
	if err == nil {
		err = p.parseNode(ctx, node)
	}

	// Spread error across the node's paths (YAML and/or SQL files)
	if err != nil {
		var pathErr pathError
		if errors.As(err, &pathErr) {
			// If there's an error in either of the YAML or SQL files, we attach a "skipped" error to the other file as well.
			for _, path := range paths {
				if path == pathErr.path {
					p.addParseError(path, err)
				} else {
					p.addParseError(path, fmt.Errorf("skipping file due to error in companion SQL/YAML file"))
				}
			}
		} else {
			// Not a path error – we add the error to all paths
			for _, path := range paths {
				p.addParseError(path, err)
			}
		}
	}

	return nil
}

// inferUnspecifiedRefs populates r.Refs with a) all explicit refs from r.rawRefs, and b) any implicit refs that we can infer from context.
// An implicit ref is one where the kind is unspecified. They are common when extracted from SQL.
// For example, if a model contains "SELECT * FROM foo", we add "foo" to r.rawRefs, and need to infer whether "foo" is a source or a model.
//
// If an unspecified ref can't be matched to another resource, it is not added to r.Refs.
// That allows, for example, a model like "SELECT * FROM foo", to parse successfully even when no other model or source is named "foo" exists.
// This is necessary to support referencing existing tables in a database. Errors for such cases will be thrown from the downstream reconciliation logic instead.
// We may want to revisit this handling in the future.
func (p *Parser) inferUnspecifiedRefs(r *Resource) {
	var refs []ResourceName
	for _, ref := range r.rawRefs {
		if ref.Kind != ResourceKindUnspecified {
			refs = append(refs, ref)
			continue
		}

		// Rule 1: If it's a model and there's a source with that name, use it
		if r.Name.Kind == ResourceKindModel {
			n := ResourceName{Kind: ResourceKindSource, Name: ref.Name}
			if _, ok := p.Resources[n.Normalized()]; ok {
				refs = append(refs, n)
				continue
			}
		}

		// Rule 2: If it's a metrics view and there's a model or source with that name, use it
		if r.Name.Kind == ResourceKindMetricsView {
			n := ResourceName{Kind: ResourceKindModel, Name: ref.Name}
			if _, ok := p.Resources[n.Normalized()]; ok {
				refs = append(refs, n)
				continue
			}
			n = ResourceName{Kind: ResourceKindSource, Name: ref.Name}
			if _, ok := p.Resources[n.Normalized()]; ok {
				refs = append(refs, n)
				continue
			}
		}

		// Rule 3: If there's a resource of the same kind with that name, use it
		n := ResourceName{Kind: r.Name.Kind, Name: ref.Name}
		if _, ok := p.Resources[n.Normalized()]; ok {
			refs = append(refs, n)
			continue
		}

		// Rule 4: Skip it
	}

	r.Refs = refs
}

// upsertResource inserts or updates a resource in the parser's internal state.
// Upserting is required since both a YAML and SQL file may contribute information to the same resource.
// After calling upsertResource, the caller can modify the returned resource's spec, and should be cautious with overriding values that may have been set from another file.
func (p *Parser) upsertResource(kind ResourceKind, name string, paths []string, refs ...ResourceName) *Resource {
	// Create the resource if not already present (ensures the spec for its kind is never nil)
	rn := ResourceName{Kind: kind, Name: name}
	r, ok := p.Resources[rn.Normalized()]
	if ok {
		// Track in updatedResources, unless it's in insertedResources
		found := false
		for _, ir := range p.insertedResources {
			if ir.Name.Normalized() == rn.Normalized() {
				found = true
				break
			}
		}
		if !found {
			for _, ur := range p.updatedResources {
				if ur.Name.Normalized() == rn.Normalized() {
					found = true
					break
				}
			}
			if !found {
				p.updatedResources = append(p.updatedResources, r)
			}
		}
	} else {
		// Create new resource and track in insertedResources
		r = &Resource{Name: rn}
		p.Resources[rn.Normalized()] = r
		p.insertedResources = append(p.insertedResources, r)
		switch kind {
		case ResourceKindSource:
			r.SourceSpec = &runtimev1.SourceSpec{}
		case ResourceKindModel:
			r.ModelSpec = &runtimev1.ModelSpec{}
		case ResourceKindMetricsView:
			r.MetricsViewSpec = &runtimev1.MetricsViewSpec{}
		case ResourceKindMigration:
			r.MigrationSpec = &runtimev1.MigrationSpec{}
		default:
			panic(fmt.Errorf("unexpected resource kind: %s", kind.String()))
		}
	}

	// Index paths if not already present
	for _, path := range paths {
		found := false
		for _, p := range r.Paths {
			if p == path {
				found = true
				break
			}
		}
		if !found {
			r.Paths = append(r.Paths, path)
			p.resourcesForPath[path] = append(p.resourcesForPath[path], r)
		}
	}

	// Add refs that are not already present
	for _, refA := range refs {
		found := false
		for _, refB := range r.rawRefs {
			if refA.Normalized() == refB.Normalized() {
				found = true
				break
			}
		}
		if !found {
			// Add to r.rawRefs
			r.rawRefs = append(r.rawRefs, refA)

			// Index in p.resourcesForUnspecifiedRef
			if refA.Kind == ResourceKindUnspecified {
				n := strings.ToLower(refA.Name)
				p.resourcesForUnspecifiedRef[n] = append(p.resourcesForUnspecifiedRef[n], r)
			}
		}
	}

	return r
}

// deleteResource removes a resource from p.Resources as well as all internal indexes.
func (p *Parser) deleteResource(r *Resource) {
	// Remove from p.Resources
	delete(p.Resources, r.Name.Normalized())

	// Remove from p.insertedResources
	checkUpdatedResources := true
	idx := slices.Index(p.insertedResources, r)
	if idx >= 0 {
		p.insertedResources = slices.Delete(p.insertedResources, idx, idx+1)
		checkUpdatedResources = false
	}

	// Remove from p.updatedResources
	idx = slices.Index(p.updatedResources, r)
	if checkUpdatedResources && idx >= 0 {
		p.updatedResources = slices.Delete(p.updatedResources, idx, idx+1)
	}

	// Remove from p.resourcesForPath
	for _, path := range r.Paths {
		rs := p.resourcesForPath[path]
		idx := slices.Index(rs, r)
		if idx < 0 {
			panic(fmt.Errorf("resource %q not found in resourcesForPath", r))
		}
		if len(rs) == 1 {
			delete(p.resourcesForPath, path)
		} else {
			p.resourcesForPath[path] = slices.Delete(rs, idx, idx+1)
		}
	}

	// Remove pointers indexed in resourcesForUnspecifiedRef
	for _, ref := range r.rawRefs {
		if ref.Kind != ResourceKindUnspecified {
			continue
		}
		n := strings.ToLower(ref.Name)
		rs := p.resourcesForUnspecifiedRef[n]
		idx := slices.Index(rs, r)
		if idx < 0 {
			panic(fmt.Errorf("resource %q not found in resourcesForUnspecifiedRef for ref %q", r.Name, ref.Name))
		}
		if len(rs) == 1 {
			delete(p.resourcesForUnspecifiedRef, n)
		} else {
			p.resourcesForUnspecifiedRef[n] = slices.Delete(rs, idx, idx+1)
		}
	}
}

// replaceResourceWithError removes a resource from the parser's internal state and adds a parse error for its paths instead.
func (p *Parser) replaceResourceWithError(n ResourceName, err error) {
	r := p.Resources[n.Normalized()]
	p.deleteResource(r)
	for _, path := range r.Paths {
		p.addParseError(path, err)
	}
}

// addParseError adds a parse error to the p.Errors
func (p *Parser) addParseError(path string, err error) {
	var loc *runtimev1.CharLocation
	var locErr locationError
	if errors.As(err, &locErr) {
		loc = locErr.location
	}

	p.Errors = append(p.Errors, &runtimev1.ParseError{
		Message:       err.Error(),
		FilePath:      path,
		StartLocation: loc,
	})
}

// normalizePath normalizes a user-provided path to the format returned from ListRecursive.
// TODO: Change this once ListRecursive returns paths without leading slash.
func normalizePath(path string) string {
	if path != "" && path[0] != '/' {
		return "/" + path
	}
	return path
}

// pathStem returns a slice of the path without the final file extension.
// If the path does not contain a file extension, the entire path is returned.f
func pathStem(path string) string {
	i := strings.LastIndexByte(path, '.')
	if i == -1 {
		return path
	}
	return path[:i]
}

// locationError wraps an error with source file character location information
type locationError struct {
	err      error
	location *runtimev1.CharLocation
}

func (e locationError) Error() string {
	return e.err.Error()
}

func (e locationError) Unwrap() error {
	return e.err
}

// pathError wraps an error with source file path information
type pathError struct {
	err  error
	path string
}

func (e pathError) Error() string {
	return e.err.Error()
}

func (e pathError) Unwrap() error {
	return e.err
}

// yamlErrLineRegexp matches the line number in a YAML error
var yamlErrLineRegexp = regexp.MustCompile(`^yaml: line (\d+):`)

// newYAMLError wraps a YAML error, extracting line number information if available
func newYAMLError(err error) error {
	res := yamlErrLineRegexp.FindStringSubmatch(err.Error())
	if len(res) != 2 {
		return err
	}

	line, err2 := strconv.Atoi(res[1])
	if err2 != nil {
		return err
	}

	return locationError{
		err: err,
		location: &runtimev1.CharLocation{
			Line: uint32(line),
		},
	}
}

// duckDBErrLineRegexp matches the line number in a DuckDB parser error
var duckDBErrLineRegexp = regexp.MustCompile(`\nLINE (\d+):`)

// newDuckDBError wraps a DuckDB parser error, extracting line number information if available
func newDuckDBError(err error) error {
	res := duckDBErrLineRegexp.FindStringSubmatch(err.Error())
	if len(res) != 2 {
		return err
	}

	line, err2 := strconv.Atoi(res[1])
	if err2 != nil {
		return err
	}

	return locationError{
		err: err,
		location: &runtimev1.CharLocation{
			Line: uint32(line),
		},
	}
}
