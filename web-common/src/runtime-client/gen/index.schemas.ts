/**
 * Generated by orval v6.10.1 🍺
 * Do not edit manually.
 * runtime.proto
 * OpenAPI spec version: version not set
 */
export type RuntimeServiceRenameFileBody = {
  fromPath?: string;
  toPath?: string;
};

export type RuntimeServicePutFileBody = {
  blob?: string;
  create?: boolean;
  /** Will cause the operation to fail if the file already exists.
It should only be set when create = true. */
  createOnly?: boolean;
};

export type RuntimeServiceListFilesParams = { glob?: string };

export type RuntimeServiceListReposParams = {
  pageSize?: number;
  pageToken?: string;
};

/**
 * Request for RuntimeService.GetTopK. Returns the top K values for a given column using agg function for table table_name.
 */
export type RuntimeServiceGetTopKBody = {
  agg?: string;
  k?: number;
};

export type RuntimeServiceTableRowsParams = { limit?: number };

export type RuntimeServiceRenameDatabaseObjectType =
  typeof RuntimeServiceRenameDatabaseObjectType[keyof typeof RuntimeServiceRenameDatabaseObjectType];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const RuntimeServiceRenameDatabaseObjectType = {
  TABLE: "TABLE",
  VIEW: "VIEW",
  FUNCTION: "FUNCTION",
} as const;

export type RuntimeServiceRenameDatabaseObjectParams = {
  name?: string;
  newname?: string;
  type?: RuntimeServiceRenameDatabaseObjectType;
};

export type RuntimeServiceQueryDirectBody = {
  sql?: string;
  args?: unknown[];
  priority?: string;
  dryRun?: boolean;
};

export type RuntimeServiceQueryBody = {
  sql?: string;
  args?: unknown[];
  priority?: string;
  dryRun?: boolean;
};

export type RuntimeServiceMigrateDeleteBody = {
  name?: string;
};

export type RuntimeServiceMigrateSingleBody = {
  sql?: string;
  dryRun?: boolean;
  createOrReplace?: boolean;
  /** If provided, will attempt to rename an existing object and only recompute if necessary.
NOTE: very questionable semantics here. */
  renameFrom?: string;
};

export type RuntimeServiceMigrateBody = {
  repoId?: string;
  /** Changed paths provides a way to "hint" what files have changed in the repo, enabling
migrations to execute faster by not scanning all code artifacts for changes. */
  changedPaths?: string[];
  dry?: boolean;
  strict?: boolean;
};

export type RuntimeServiceMetricsViewTotalsBody = {
  measureNames?: string[];
  timeStart?: string;
  timeEnd?: string;
  filter?: V1MetricsViewFilter;
};

export type RuntimeServiceMetricsViewToplistBody = {
  measureNames?: string[];
  timeStart?: string;
  timeEnd?: string;
  limit?: string;
  offset?: string;
  sort?: V1MetricsViewSort[];
  filter?: V1MetricsViewFilter;
};

export type RuntimeServiceMetricsViewTimeSeriesBody = {
  measureNames?: string[];
  timeStart?: string;
  timeEnd?: string;
  timeGranularity?: string;
  filter?: V1MetricsViewFilter;
};

export type RuntimeServiceGenerateTimeSeriesBody = {
  tableName?: string;
  measures?: GenerateTimeSeriesRequestBasicMeasures;
  timestampColumnName?: string;
  timeRange?: V1TimeSeriesTimeRange;
  filters?: V1MetricsViewRequestFilter;
  pixels?: string;
  sampleSize?: number;
};

export type RuntimeServiceEstimateRollupIntervalBody = {
  columnName?: string;
};

export type RuntimeServiceListCatalogObjectsType =
  typeof RuntimeServiceListCatalogObjectsType[keyof typeof RuntimeServiceListCatalogObjectsType];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const RuntimeServiceListCatalogObjectsType = {
  TYPE_UNSPECIFIED: "TYPE_UNSPECIFIED",
  TYPE_TABLE: "TYPE_TABLE",
  TYPE_SOURCE: "TYPE_SOURCE",
  TYPE_MODEL: "TYPE_MODEL",
  TYPE_METRICS_VIEW: "TYPE_METRICS_VIEW",
} as const;

export type RuntimeServiceListCatalogObjectsParams = {
  type?: RuntimeServiceListCatalogObjectsType;
};

export type RuntimeServiceListInstancesParams = {
  pageSize?: number;
  pageToken?: string;
};

export type V1TypeCode = typeof V1TypeCode[keyof typeof V1TypeCode];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1TypeCode = {
  CODE_UNSPECIFIED: "CODE_UNSPECIFIED",
  CODE_BOOL: "CODE_BOOL",
  CODE_INT8: "CODE_INT8",
  CODE_INT16: "CODE_INT16",
  CODE_INT32: "CODE_INT32",
  CODE_INT64: "CODE_INT64",
  CODE_INT128: "CODE_INT128",
  CODE_UINT8: "CODE_UINT8",
  CODE_UINT16: "CODE_UINT16",
  CODE_UINT32: "CODE_UINT32",
  CODE_UINT64: "CODE_UINT64",
  CODE_UINT128: "CODE_UINT128",
  CODE_FLOAT32: "CODE_FLOAT32",
  CODE_FLOAT64: "CODE_FLOAT64",
  CODE_TIMESTAMP: "CODE_TIMESTAMP",
  CODE_DATE: "CODE_DATE",
  CODE_TIME: "CODE_TIME",
  CODE_STRING: "CODE_STRING",
  CODE_BYTES: "CODE_BYTES",
  CODE_ARRAY: "CODE_ARRAY",
  CODE_STRUCT: "CODE_STRUCT",
  CODE_MAP: "CODE_MAP",
  CODE_DECIMAL: "CODE_DECIMAL",
  CODE_JSON: "CODE_JSON",
  CODE_UUID: "CODE_UUID",
} as const;

export interface V1TriggerSyncResponse {
  objectsCount?: number;
  objectsAddedCount?: number;
  objectsUpdatedCount?: number;
  objectsRemovedCount?: number;
}

export interface V1TriggerRefreshResponse {
  [key: string]: any;
}

/**
 * Response for RuntimeService.GetTopK.
 */
export interface V1TopKResponse {
  entries?: TopKResponseTopKEntry[];
}

export type V1TimeSeriesValueRecords = { [key: string]: number };

export interface V1TimeSeriesValue {
  ts?: string;
  bin?: number;
  records?: V1TimeSeriesValueRecords;
}

export interface V1TimeSeriesTimeRange {
  name?: V1TimeRangeName;
  start?: string;
  end?: string;
  interval?: string;
}

export interface V1TimeSeriesResponse {
  id?: string;
  results?: V1TimeSeriesValue[];
  spark?: TimeSeriesResponseTimeSeriesValues;
  timeRange?: V1TimeSeriesTimeRange;
  sampleSize?: number;
  error?: string;
}

export interface V1TimeSeriesRollup {
  rollup?: V1TimeSeriesResponse;
}

export interface V1TimeRangeSummary {
  min?: string;
  max?: string;
  interval?: TimeRangeSummaryInterval;
}

export type V1TimeRangeName =
  typeof V1TimeRangeName[keyof typeof V1TimeRangeName];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1TimeRangeName = {
  LastHour: "LastHour",
  Last6Hours: "Last6Hours",
  LastDay: "LastDay",
  Last2Days: "Last2Days",
  Last5Days: "Last5Days",
  LastWeek: "LastWeek",
  Last2Weeks: "Last2Weeks",
  Last30Days: "Last30Days",
  Last60Days: "Last60Days",
  AllTime: "AllTime",
} as const;

export type V1TimeGrain = typeof V1TimeGrain[keyof typeof V1TimeGrain];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1TimeGrain = {
  MILLISECOND: "MILLISECOND",
  SECOND: "SECOND",
  MINUTE: "MINUTE",
  HOUR: "HOUR",
  DAY: "DAY",
  WEEK: "WEEK",
  MONTH: "MONTH",
  YEAR: "YEAR",
  UNSPECIFIED: "UNSPECIFIED",
} as const;

export interface V1TimeSeriesTimeRange {
  name?: V1TimeRangeName;
  start?: string;
  end?: string;
  interval?: string;
}

export interface V1StructType {
  fields?: StructTypeField[];
}

/**
 * Table represents a table in the OLAP database. These include pre-existing tables discovered by periodically
scanning the database's information schema when the instance is created with exposed=true. Pre-existing tables
have managed = false.
 */
export interface V1Table {
  name?: string;
  schema?: V1StructType;
  /** Managed is true if the table was created through a runtime migration, false if it was discovered in by
scanning the database's information schema. */
  managed?: boolean;
}

export type V1SourceProperties = { [key: string]: any };

export interface V1Source {
  name?: string;
  connector?: string;
  properties?: V1SourceProperties;
  schema?: V1StructType;
  sql?: string;
}

export interface V1Scalar {
  int64?: string;
  double?: number;
  timestamp?: string;
}

export type V1RowsResponseDataItem = { [key: string]: any };

export interface V1RowsResponse {
  data?: V1RowsResponseDataItem[];
}

/**
 * Repo represents a collection of file artifacts containing SQL statements.
It will usually by represented as a folder on disk, but may also be backed by a
database (for modelling in the cloud where no persistant file system is available).
 */
export interface V1Repo {
  repoId?: string;
  /** Driver for persisting artifacts. Supports "file" and "postgres". */
  driver?: string;
  /** DSN for driver. If the driver is "file", this should be the path to the root directory. */
  dsn?: string;
}

export interface V1RenameFileResponse {
  [key: string]: any;
}

export interface V1RenameFileAndMigrateResponse {
  /** Errors encountered during the migration. If strict = false, any path in
affected_paths without an error can be assumed to have been migrated succesfully. */
  errors?: V1MigrationError[];
  /** affected_paths lists all the file paths that were considered while
executing the migration. For a PutFileAndMigrate, this includes the put file
as well as any file artifacts that rely on objects declared in it. */
  affectedPaths?: string[];
}

export interface V1RenameFileAndMigrateRequest {
  repoId?: string;
  instanceId?: string;
  fromPath?: string;
  toPath?: string;
  /** If true, will save the file and validate it and related file artifacts, but not actually execute any migrations. */
  dry?: boolean;
  strict?: boolean;
}

export interface V1RenameDatabaseObjectResponse {
  [key: string]: any;
}

export type V1QueryResponseDataItem = { [key: string]: any };

export interface V1QueryResponse {
  meta?: V1StructType;
  data?: V1QueryResponseDataItem[];
}

export type V1QueryDirectResponseDataItem = { [key: string]: any };

export interface V1QueryDirectResponse {
  meta?: V1StructType;
  data?: V1QueryDirectResponseDataItem[];
}

export interface V1PutFileResponse {
  filePath?: string;
}

export interface V1PutFileAndMigrateResponse {
  /** Errors encountered during the migration. If strict = false, any path in
affected_paths without an error can be assumed to have been migrated succesfully. */
  errors?: V1MigrationError[];
  /** affected_paths lists all the file paths that were considered while
executing the migration. For a PutFileAndMigrate, this includes the put file
as well as any file artifacts that rely on objects declared in it. */
  affectedPaths?: string[];
}

export interface V1PutFileAndMigrateRequest {
  repoId?: string;
  instanceId?: string;
  path?: string;
  blob?: string;
  create?: boolean;
  /** create_only will cause the operation to fail if a file already exists at path.
It should only be set when create = true. */
  createOnly?: boolean;
  /** If true, will save the file and validate it and related file artifacts, but not actually execute any migrations. */
  dry?: boolean;
  strict?: boolean;
}

export interface V1ProfileColumn {
  name?: string;
  type?: string;
  largestStringLength?: number;
}

export interface V1ProfileColumnsResponse {
  profileColumns?: V1ProfileColumn[];
}

export interface V1PingResponse {
  version?: string;
  time?: string;
}

export interface V1NumericStatistics {
  min?: number;
  max?: number;
  mean?: number;
  q25?: number;
  q50?: number;
  q75?: number;
  sd?: number;
}

export interface V1NumericOutliers {
  outliers?: NumericOutliersOutlier[];
}

export interface V1NumericHistogramBins {
  bins?: NumericHistogramBinsBin[];
}

/**
 * Response for RuntimeService.GetNumericHistogram, RuntimeService.GetDescriptiveStatistics and RuntimeService.GetCardinalityOfColumn.
Message will have either numericHistogramBins, numericStatistics or numericOutliers set.
 */
export interface V1NumericSummary {
  numericHistogramBins?: V1NumericHistogramBins;
  numericStatistics?: V1NumericStatistics;
  numericOutliers?: V1NumericOutliers;
}

export interface V1NullCountResponse {
  count?: string;
}

export interface V1Model {
  name?: string;
  sql?: string;
  dialect?: ModelDialect;
  schema?: V1StructType;
}

/**
 * - CODE_UNSPECIFIED: Unspecified error
 - CODE_SYNTAX: Code artifact failed to parse
 - CODE_VALIDATION: Code artifact has internal validation errors
 - CODE_DEPENDENCY: Code artifact is valid, but has invalid dependencies
 - CODE_OLAP: Error returned from the OLAP database
 - CODE_SOURCE: Error encountered during source inspection or ingestion
 */
export type V1MigrationErrorCode =
  typeof V1MigrationErrorCode[keyof typeof V1MigrationErrorCode];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1MigrationErrorCode = {
  CODE_UNSPECIFIED: "CODE_UNSPECIFIED",
  CODE_SYNTAX: "CODE_SYNTAX",
  CODE_VALIDATION: "CODE_VALIDATION",
  CODE_DEPENDENCY: "CODE_DEPENDENCY",
  CODE_OLAP: "CODE_OLAP",
  CODE_SOURCE: "CODE_SOURCE",
} as const;

export interface V1MigrateSingleResponse {
  [key: string]: any;
}

export interface V1MigrateResponse {
  /** Errors encountered during the migration. If strict = false, any path in
affected_paths without an error can be assumed to have been migrated succesfully. */
  errors?: V1MigrationError[];
  /** affected_paths lists all the file artifact paths that were considered while
executing the migration. If changed_paths was empty, this will include all
code artifacts in the repo. */
  affectedPaths?: string[];
}

export interface V1MigrateDeleteResponse {
  [key: string]: any;
}

export type V1MetricsViewTotalsResponseData = { [key: string]: any };

export interface V1MetricsViewTotalsResponse {
  meta?: V1MetricsViewColumn[];
  data?: V1MetricsViewTotalsResponseData;
}

export type V1MetricsViewToplistResponseDataItem = { [key: string]: any };

export interface V1MetricsViewToplistResponse {
  meta?: V1MetricsViewColumn[];
  data?: V1MetricsViewToplistResponseDataItem[];
}

export type V1MetricsViewTimeSeriesResponseDataItem = { [key: string]: any };

export interface V1MetricsViewTimeSeriesResponse {
  meta?: V1MetricsViewColumn[];
  data?: V1MetricsViewTimeSeriesResponseDataItem[];
}

export interface V1MetricsViewSort {
  name?: string;
  ascending?: boolean;
}

export interface V1MetricsViewRequestFilter {
  include?: V1MetricsViewDimensionValue[];
  exclude?: V1MetricsViewDimensionValue[];
}

export interface V1MetricsViewMetaResponse {
  metricsViewName?: string;
  fromObject?: string;
  dimensions?: MetricsViewDimension[];
  measures?: MetricsViewMeasure[];
}

export interface V1MetricsViewFilter {
  match?: string[];
  include?: MetricsViewFilterCond[];
  exclude?: MetricsViewFilterCond[];
}

export interface V1MetricsViewDimensionValue {
  name?: string;
  in?: unknown[];
  like?: MetricsViewDimensionValueValues;
}

export interface V1MetricsViewColumn {
  name?: string;
  type?: string;
  nullable?: boolean;
}

export interface V1MetricsView {
  name?: string;
  from?: string;
  timeDimension?: string;
  /** Recommended granularities for rolling up the time dimension.
Should be a valid SQL INTERVAL value. */
  timeGrains?: string[];
  dimensions?: MetricsViewDimension[];
  measures?: MetricsViewMeasure[];
}

export interface V1MapType {
  keyType?: Runtimev1Type;
  valueType?: Runtimev1Type;
}

export interface V1ListReposResponse {
  repos?: V1Repo[];
  nextPageToken?: string;
}

export interface V1ListInstancesResponse {
  instances?: V1Instance[];
  nextPageToken?: string;
}

export interface V1ListFilesResponse {
  paths?: string[];
}

export interface V1ListConnectorsResponse {
  connectors?: V1Connector[];
}

export interface V1ListCatalogObjectsResponse {
  objects?: V1CatalogObject[];
}

/**
 * Instance represents one connection to an OLAP datastore (such as DuckDB or Druid).
Migrations and queries are issued against a specific instance. The concept of
instances enables multiple data projects to be served by one runtime.
 */
export interface V1Instance {
  instanceId?: string;
  driver?: string;
  dsn?: string;
  /** Prefix to add to all table names created through Rill SQL (such as sources, models, etc.)
Use it as an alternative to database schemas. */
  objectPrefix?: string;
  /** Indicates that the underlying infra may be manipulated directly by users.
If true, the runtime will continuously poll the infra's information schema
to discover tables not created through the runtime. They will be added to the
catalog as UnmanagedTables. */
  exposed?: boolean;
  /** If true, the runtime will store the instance's catalog data (such as sources and metrics views)
in the instance's OLAP datastore instead of in the runtime's metadata store. This is currently
only supported for the duckdb driver. */
  embedCatalog?: boolean;
}

export interface V1GetRepoResponse {
  repo?: V1Repo;
}

export interface V1GetInstanceResponse {
  instance?: V1Instance;
}

export interface V1GetFileResponse {
  blob?: string;
  updatedOn?: string;
}

export interface V1GetCatalogObjectResponse {
  object?: V1CatalogObject;
}

export interface V1EstimateSmallestTimeGrainResponse {
  timeGrain?: V1TimeGrain;
}

export interface V1EstimateRollupIntervalResponse {
  interval?: string;
  min?: V1Scalar;
  max?: V1Scalar;
}

export interface V1DeleteRepoResponse {
  [key: string]: any;
}

export interface V1DeleteInstanceResponse {
  [key: string]: any;
}

export interface V1DeleteFileResponse {
  [key: string]: any;
}

export interface V1DeleteFileAndMigrateResponse {
  /** Errors encountered during the migration. If strict = false, any path in
affected_paths without an error can be assumed to have been migrated succesfully. */
  errors?: V1MigrationError[];
  /** affected_paths lists all the file paths that were considered while
executing the migration. For a PutFileAndMigrate, this includes the put file
as well as any file artifacts that rely on objects declared in it. */
  affectedPaths?: string[];
}

export interface V1DeleteFileAndMigrateRequest {
  repoId?: string;
  instanceId?: string;
  path?: string;
  /** If true, will save the file and validate it and related file artifacts, but not actually execute any migrations. */
  dry?: boolean;
  strict?: boolean;
}

export type V1DatabaseObjectType =
  typeof V1DatabaseObjectType[keyof typeof V1DatabaseObjectType];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1DatabaseObjectType = {
  TABLE: "TABLE",
  VIEW: "VIEW",
  FUNCTION: "FUNCTION",
} as const;

export interface V1CreateRepoResponse {
  repo?: V1Repo;
}

export interface V1CreateRepoRequest {
  repoId?: string;
  driver?: string;
  dsn?: string;
}

export interface V1CreateInstanceResponse {
  instanceId?: string;
  instance?: V1Instance;
}

export interface V1CreateInstanceRequest {
  instanceId?: string;
  driver?: string;
  dsn?: string;
  objectPrefix?: string;
  exposed?: boolean;
  embedCatalog?: boolean;
}

/**
 * Connector represents a connector available in the runtime.
It should not be confused with a source.
 */
export interface V1Connector {
  name?: string;
  displayName?: string;
  description?: string;
  properties?: ConnectorProperty[];
}

/**
 * Response for RuntimeService.GetTopK and RuntimeService.GetCardinalityOfColumn. Message will have either topK or cardinality set.
 */
export interface V1CategoricalSummary {
  topKResponse?: V1TopKResponse;
  cardinality?: string;
}

export type V1CatalogObjectType =
  typeof V1CatalogObjectType[keyof typeof V1CatalogObjectType];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const V1CatalogObjectType = {
  TYPE_UNSPECIFIED: "TYPE_UNSPECIFIED",
  TYPE_TABLE: "TYPE_TABLE",
  TYPE_SOURCE: "TYPE_SOURCE",
  TYPE_MODEL: "TYPE_MODEL",
  TYPE_METRICS_VIEW: "TYPE_METRICS_VIEW",
} as const;

export interface V1CatalogObject {
  type?: V1CatalogObjectType;
  table?: V1Table;
  source?: V1Source;
  model?: V1Model;
  metricsView?: V1MetricsView;
  name?: string;
  path?: string;
  createdOn?: string;
  updatedOn?: string;
  refreshedOn?: string;
}

export interface V1CardinalityResponse {
  cardinality?: string;
}

export interface V1BasicMeasureDefinition {
  id?: string;
  expression?: string;
  sqlName?: string;
}

export interface Runtimev1Type {
  code?: V1TypeCode;
  nullable?: boolean;
  arrayElementType?: Runtimev1Type;
  structType?: V1StructType;
  mapType?: V1MapType;
}

export interface RpcStatus {
  code?: number;
  message?: string;
  details?: ProtobufAny[];
}

/**
 * `NullValue` is a singleton enumeration to represent the null value for the
`Value` type union.

 The JSON representation for `NullValue` is JSON `null`.

 - NULL_VALUE: Null value.
 */
export type ProtobufNullValue =
  typeof ProtobufNullValue[keyof typeof ProtobufNullValue];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const ProtobufNullValue = {
  NULL_VALUE: "NULL_VALUE",
} as const;

export interface ProtobufAny {
  "@type"?: string;
  [key: string]: unknown;
}

export interface TopKResponseTopKEntry {
  /** value is optional so that null values from the database can be represented. */
  value?: string;
  count?: number;
}

export interface TimeSeriesResponseTimeSeriesValues {
  values?: V1TimeSeriesValue[];
}

export interface TimeRangeSummaryInterval {
  months?: number;
  days?: number;
  micros?: string;
}

export interface StructTypeField {
  name?: string;
  type?: Runtimev1Type;
}

export interface NumericOutliersOutlier {
  bucket?: string;
  low?: number;
  high?: number;
  present?: boolean;
}

export interface NumericHistogramBinsBin {
  bucket?: string;
  low?: number;
  high?: number;
  count?: string;
}

export type ModelDialect = typeof ModelDialect[keyof typeof ModelDialect];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const ModelDialect = {
  DIALECT_UNSPECIFIED: "DIALECT_UNSPECIFIED",
  DIALECT_DUCKDB: "DIALECT_DUCKDB",
} as const;

export interface MigrationErrorCharLocation {
  line?: number;
  column?: number;
}

/**
 * MigrationError represents an error encountered while running Migrate.
 */
export interface V1MigrationError {
  code?: V1MigrationErrorCode;
  message?: string;
  filePath?: string;
  /** Property path of the error in the code artifact (if any).
It's represented as a JS-style property path, e.g. "key0.key1[index2].key3".
It only applies to structured code artifacts (i.e. YAML).
Only applicable if file_path is set. */
  propertyPath?: string;
  startLocation?: MigrationErrorCharLocation;
  endLocation?: MigrationErrorCharLocation;
}

export interface MetricsViewMeasure {
  name?: string;
  label?: string;
  expression?: string;
  description?: string;
  format?: string;
  enabled?: string;
}

export interface MetricsViewFilterCond {
  name?: string;
  in?: unknown[];
  like?: unknown[];
}

export interface MetricsViewDimensionValueValues {
  values?: unknown[];
}

export interface MetricsViewDimension {
  name?: string;
  label?: string;
  description?: string;
  enabled?: string;
}

export interface GenerateTimeSeriesRequestBasicMeasures {
  basicMeasures?: V1BasicMeasureDefinition[];
}

export type ConnectorPropertyType =
  typeof ConnectorPropertyType[keyof typeof ConnectorPropertyType];

// eslint-disable-next-line @typescript-eslint/no-redeclare
export const ConnectorPropertyType = {
  TYPE_UNSPECIFIED: "TYPE_UNSPECIFIED",
  TYPE_STRING: "TYPE_STRING",
  TYPE_NUMBER: "TYPE_NUMBER",
  TYPE_BOOLEAN: "TYPE_BOOLEAN",
  TYPE_INFORMATIONAL: "TYPE_INFORMATIONAL",
} as const;

export interface ConnectorProperty {
  key?: string;
  displayName?: string;
  description?: string;
  placeholder?: string;
  type?: ConnectorPropertyType;
  nullable?: boolean;
  hint?: string;
  href?: string;
}
