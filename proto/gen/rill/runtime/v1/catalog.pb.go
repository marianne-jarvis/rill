// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: rill/runtime/v1/catalog.proto

package runtimev1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// ObjectType represents the different kinds of catalog objects
type ObjectType int32

const (
	ObjectType_OBJECT_TYPE_UNSPECIFIED  ObjectType = 0
	ObjectType_OBJECT_TYPE_TABLE        ObjectType = 1
	ObjectType_OBJECT_TYPE_SOURCE       ObjectType = 2
	ObjectType_OBJECT_TYPE_MODEL        ObjectType = 3
	ObjectType_OBJECT_TYPE_METRICS_VIEW ObjectType = 4
)

// Enum value maps for ObjectType.
var (
	ObjectType_name = map[int32]string{
		0: "OBJECT_TYPE_UNSPECIFIED",
		1: "OBJECT_TYPE_TABLE",
		2: "OBJECT_TYPE_SOURCE",
		3: "OBJECT_TYPE_MODEL",
		4: "OBJECT_TYPE_METRICS_VIEW",
	}
	ObjectType_value = map[string]int32{
		"OBJECT_TYPE_UNSPECIFIED":  0,
		"OBJECT_TYPE_TABLE":        1,
		"OBJECT_TYPE_SOURCE":       2,
		"OBJECT_TYPE_MODEL":        3,
		"OBJECT_TYPE_METRICS_VIEW": 4,
	}
)

func (x ObjectType) Enum() *ObjectType {
	p := new(ObjectType)
	*p = x
	return p
}

func (x ObjectType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ObjectType) Descriptor() protoreflect.EnumDescriptor {
	return file_rill_runtime_v1_catalog_proto_enumTypes[0].Descriptor()
}

func (ObjectType) Type() protoreflect.EnumType {
	return &file_rill_runtime_v1_catalog_proto_enumTypes[0]
}

func (x ObjectType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ObjectType.Descriptor instead.
func (ObjectType) EnumDescriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{0}
}

// Dialects supported for models
type Model_Dialect int32

const (
	Model_DIALECT_UNSPECIFIED Model_Dialect = 0
	Model_DIALECT_DUCKDB      Model_Dialect = 1
)

// Enum value maps for Model_Dialect.
var (
	Model_Dialect_name = map[int32]string{
		0: "DIALECT_UNSPECIFIED",
		1: "DIALECT_DUCKDB",
	}
	Model_Dialect_value = map[string]int32{
		"DIALECT_UNSPECIFIED": 0,
		"DIALECT_DUCKDB":      1,
	}
)

func (x Model_Dialect) Enum() *Model_Dialect {
	p := new(Model_Dialect)
	*p = x
	return p
}

func (x Model_Dialect) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Model_Dialect) Descriptor() protoreflect.EnumDescriptor {
	return file_rill_runtime_v1_catalog_proto_enumTypes[1].Descriptor()
}

func (Model_Dialect) Type() protoreflect.EnumType {
	return &file_rill_runtime_v1_catalog_proto_enumTypes[1]
}

func (x Model_Dialect) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Model_Dialect.Descriptor instead.
func (Model_Dialect) EnumDescriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{2, 0}
}

// Table represents a table in the OLAP database. These include pre-existing tables discovered by periodically
// scanning the database's information schema when the instance is created with exposed=true. Pre-existing tables
// have managed = false.
type Table struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Table name
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Table schema
	Schema *StructType `protobuf:"bytes,2,opt,name=schema,proto3" json:"schema,omitempty"`
	// Managed is true if the table was created through a runtime migration, false if it was discovered in by
	// scanning the database's information schema.
	Managed bool `protobuf:"varint,3,opt,name=managed,proto3" json:"managed,omitempty"`
}

func (x *Table) Reset() {
	*x = Table{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_catalog_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Table) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Table) ProtoMessage() {}

func (x *Table) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_catalog_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Table.ProtoReflect.Descriptor instead.
func (*Table) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{0}
}

func (x *Table) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Table) GetSchema() *StructType {
	if x != nil {
		return x.Schema
	}
	return nil
}

func (x *Table) GetManaged() bool {
	if x != nil {
		return x.Managed
	}
	return false
}

// Source is the internal representation of a source definition
type Source struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the source
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Connector used by the source
	Connector string `protobuf:"bytes,2,opt,name=connector,proto3" json:"connector,omitempty"`
	// Connector properties assigned in the source
	Properties *structpb.Struct `protobuf:"bytes,3,opt,name=properties,proto3" json:"properties,omitempty"`
	// Detected schema of the source
	Schema *StructType `protobuf:"bytes,5,opt,name=schema,proto3" json:"schema,omitempty"`
	// extraction policy for the source
	Policy *Source_ExtractPolicy `protobuf:"bytes,6,opt,name=policy,proto3" json:"policy,omitempty"`
}

func (x *Source) Reset() {
	*x = Source{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_catalog_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Source) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Source) ProtoMessage() {}

func (x *Source) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_catalog_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Source.ProtoReflect.Descriptor instead.
func (*Source) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{1}
}

func (x *Source) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Source) GetConnector() string {
	if x != nil {
		return x.Connector
	}
	return ""
}

func (x *Source) GetProperties() *structpb.Struct {
	if x != nil {
		return x.Properties
	}
	return nil
}

func (x *Source) GetSchema() *StructType {
	if x != nil {
		return x.Schema
	}
	return nil
}

func (x *Source) GetPolicy() *Source_ExtractPolicy {
	if x != nil {
		return x.Policy
	}
	return nil
}

// Model is the internal representation of a model definition
type Model struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the model
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// SQL is a SELECT statement representing the model
	Sql string `protobuf:"bytes,2,opt,name=sql,proto3" json:"sql,omitempty"`
	// Dialect of the SQL statement
	Dialect Model_Dialect `protobuf:"varint,3,opt,name=dialect,proto3,enum=rill.runtime.v1.Model_Dialect" json:"dialect,omitempty"`
	// Detected schema of the model
	Schema *StructType `protobuf:"bytes,4,opt,name=schema,proto3" json:"schema,omitempty"`
	// To materialize model or not
	Materialize bool `protobuf:"varint,5,opt,name=materialize,proto3" json:"materialize,omitempty"`
}

func (x *Model) Reset() {
	*x = Model{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_catalog_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Model) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Model) ProtoMessage() {}

func (x *Model) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_catalog_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Model.ProtoReflect.Descriptor instead.
func (*Model) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{2}
}

func (x *Model) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Model) GetSql() string {
	if x != nil {
		return x.Sql
	}
	return ""
}

func (x *Model) GetDialect() Model_Dialect {
	if x != nil {
		return x.Dialect
	}
	return Model_DIALECT_UNSPECIFIED
}

func (x *Model) GetSchema() *StructType {
	if x != nil {
		return x.Schema
	}
	return nil
}

func (x *Model) GetMaterialize() bool {
	if x != nil {
		return x.Materialize
	}
	return false
}

// Metrics view is the internal representation of a metrics view definition
type MetricsView struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Name of the metrics view
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Name of the source or model that the metrics view is based on
	Model string `protobuf:"bytes,2,opt,name=model,proto3" json:"model,omitempty"`
	// Name of the primary time dimension, used for rendering time series
	TimeDimension string `protobuf:"bytes,3,opt,name=time_dimension,json=timeDimension,proto3" json:"time_dimension,omitempty"`
	// Recommended granularities for rolling up the time dimension.
	// Should be a valid SQL INTERVAL value.
	TimeGrains []string `protobuf:"bytes,4,rep,name=time_grains,json=timeGrains,proto3" json:"time_grains,omitempty"`
	// Dimensions in the metrics view
	Dimensions []*MetricsView_Dimension `protobuf:"bytes,5,rep,name=dimensions,proto3" json:"dimensions,omitempty"`
	// Measures in the metrics view
	Measures []*MetricsView_Measure `protobuf:"bytes,6,rep,name=measures,proto3" json:"measures,omitempty"`
	// User friendly label for the dashboard
	Label string `protobuf:"bytes,7,opt,name=label,proto3" json:"label,omitempty"`
	// Brief description of the dashboard
	Description string `protobuf:"bytes,8,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *MetricsView) Reset() {
	*x = MetricsView{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_catalog_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricsView) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricsView) ProtoMessage() {}

func (x *MetricsView) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_catalog_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricsView.ProtoReflect.Descriptor instead.
func (*MetricsView) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{3}
}

func (x *MetricsView) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MetricsView) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *MetricsView) GetTimeDimension() string {
	if x != nil {
		return x.TimeDimension
	}
	return ""
}

func (x *MetricsView) GetTimeGrains() []string {
	if x != nil {
		return x.TimeGrains
	}
	return nil
}

func (x *MetricsView) GetDimensions() []*MetricsView_Dimension {
	if x != nil {
		return x.Dimensions
	}
	return nil
}

func (x *MetricsView) GetMeasures() []*MetricsView_Measure {
	if x != nil {
		return x.Measures
	}
	return nil
}

func (x *MetricsView) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *MetricsView) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

// Extract policy for glob connectors
type Source_ExtractPolicy struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Row policy to select rows
	Row *Source_ExtractPolicy_ExtractConfig `protobuf:"bytes,1,opt,name=row,proto3" json:"row,omitempty"`
	// File policy to select files
	File *Source_ExtractPolicy_ExtractConfig `protobuf:"bytes,2,opt,name=file,proto3" json:"file,omitempty"`
}

func (x *Source_ExtractPolicy) Reset() {
	*x = Source_ExtractPolicy{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_catalog_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Source_ExtractPolicy) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Source_ExtractPolicy) ProtoMessage() {}

func (x *Source_ExtractPolicy) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_catalog_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Source_ExtractPolicy.ProtoReflect.Descriptor instead.
func (*Source_ExtractPolicy) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{1, 0}
}

func (x *Source_ExtractPolicy) GetRow() *Source_ExtractPolicy_ExtractConfig {
	if x != nil {
		return x.Row
	}
	return nil
}

func (x *Source_ExtractPolicy) GetFile() *Source_ExtractPolicy_ExtractConfig {
	if x != nil {
		return x.File
	}
	return nil
}

// extract configs
type Source_ExtractPolicy_ExtractConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// strategy of the policy
	Strategy string `protobuf:"bytes,1,opt,name=strategy,proto3" json:"strategy,omitempty"`
	// size
	Size string `protobuf:"bytes,2,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *Source_ExtractPolicy_ExtractConfig) Reset() {
	*x = Source_ExtractPolicy_ExtractConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_catalog_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Source_ExtractPolicy_ExtractConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Source_ExtractPolicy_ExtractConfig) ProtoMessage() {}

func (x *Source_ExtractPolicy_ExtractConfig) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_catalog_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Source_ExtractPolicy_ExtractConfig.ProtoReflect.Descriptor instead.
func (*Source_ExtractPolicy_ExtractConfig) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{1, 0, 0}
}

func (x *Source_ExtractPolicy_ExtractConfig) GetStrategy() string {
	if x != nil {
		return x.Strategy
	}
	return ""
}

func (x *Source_ExtractPolicy_ExtractConfig) GetSize() string {
	if x != nil {
		return x.Size
	}
	return ""
}

// Dimensions are columns to filter and group by
type MetricsView_Dimension struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Label       string `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
}

func (x *MetricsView_Dimension) Reset() {
	*x = MetricsView_Dimension{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_catalog_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricsView_Dimension) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricsView_Dimension) ProtoMessage() {}

func (x *MetricsView_Dimension) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_catalog_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricsView_Dimension.ProtoReflect.Descriptor instead.
func (*MetricsView_Dimension) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{3, 0}
}

func (x *MetricsView_Dimension) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MetricsView_Dimension) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *MetricsView_Dimension) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

// Measures are aggregated computed values
type MetricsView_Measure struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name        string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Label       string `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	Expression  string `protobuf:"bytes,3,opt,name=expression,proto3" json:"expression,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Format      string `protobuf:"bytes,5,opt,name=format,proto3" json:"format,omitempty"`
}

func (x *MetricsView_Measure) Reset() {
	*x = MetricsView_Measure{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_catalog_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricsView_Measure) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricsView_Measure) ProtoMessage() {}

func (x *MetricsView_Measure) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_catalog_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricsView_Measure.ProtoReflect.Descriptor instead.
func (*MetricsView_Measure) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_catalog_proto_rawDescGZIP(), []int{3, 1}
}

func (x *MetricsView_Measure) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *MetricsView_Measure) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *MetricsView_Measure) GetExpression() string {
	if x != nil {
		return x.Expression
	}
	return ""
}

func (x *MetricsView_Measure) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *MetricsView_Measure) GetFormat() string {
	if x != nil {
		return x.Format
	}
	return ""
}

var File_rill_runtime_v1_catalog_proto protoreflect.FileDescriptor

var file_rill_runtime_v1_catalog_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76,
	0x31, 0x2f, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0f, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c,
	0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76, 0x31, 0x2f,
	0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x6a, 0x0a, 0x05,
	0x54, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x33, 0x0a, 0x06, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x72, 0x69, 0x6c, 0x6c,
	0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x18,
	0x0a, 0x07, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x07, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x64, 0x22, 0xca, 0x03, 0x0a, 0x06, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x37, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74,
	0x69, 0x65, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x69, 0x65, 0x73, 0x12, 0x33,
	0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b,
	0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x06, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x12, 0x3d, 0x0a, 0x06, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x25, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x45, 0x78, 0x74,
	0x72, 0x61, 0x63, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x06, 0x70, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x1a, 0xe0, 0x01, 0x0a, 0x0d, 0x45, 0x78, 0x74, 0x72, 0x61, 0x63, 0x74, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x12, 0x45, 0x0a, 0x03, 0x72, 0x6f, 0x77, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x33, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x45, 0x78, 0x74, 0x72, 0x61,
	0x63, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x2e, 0x45, 0x78, 0x74, 0x72, 0x61, 0x63, 0x74,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x03, 0x72, 0x6f, 0x77, 0x12, 0x47, 0x0a, 0x04, 0x66,
	0x69, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x33, 0x2e, 0x72, 0x69, 0x6c, 0x6c,
	0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x2e, 0x45, 0x78, 0x74, 0x72, 0x61, 0x63, 0x74, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79,
	0x2e, 0x45, 0x78, 0x74, 0x72, 0x61, 0x63, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x04,
	0x66, 0x69, 0x6c, 0x65, 0x1a, 0x3f, 0x0a, 0x0d, 0x45, 0x78, 0x74, 0x72, 0x61, 0x63, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x74, 0x72, 0x61, 0x74, 0x65, 0x67,
	0x79, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0xf6, 0x01, 0x0a, 0x05, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x71, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x73, 0x71, 0x6c, 0x12, 0x38, 0x0a, 0x07, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x63, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75,
	0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x44,
	0x69, 0x61, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x07, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x63, 0x74, 0x12,
	0x33, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1b, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x06, 0x73, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x12, 0x20, 0x0a, 0x0b, 0x6d, 0x61, 0x74, 0x65, 0x72, 0x69, 0x61, 0x6c,
	0x69, 0x7a, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x6d, 0x61, 0x74, 0x65, 0x72,
	0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x22, 0x36, 0x0a, 0x07, 0x44, 0x69, 0x61, 0x6c, 0x65, 0x63,
	0x74, 0x12, 0x17, 0x0a, 0x13, 0x44, 0x49, 0x41, 0x4c, 0x45, 0x43, 0x54, 0x5f, 0x55, 0x4e, 0x53,
	0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x12, 0x0a, 0x0e, 0x44, 0x49,
	0x41, 0x4c, 0x45, 0x43, 0x54, 0x5f, 0x44, 0x55, 0x43, 0x4b, 0x44, 0x42, 0x10, 0x01, 0x22, 0xaa,
	0x04, 0x0a, 0x0b, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x56, 0x69, 0x65, 0x77, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x25, 0x0a, 0x0e, 0x74, 0x69, 0x6d, 0x65,
	0x5f, 0x64, 0x69, 0x6d, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x74, 0x69, 0x6d, 0x65, 0x44, 0x69, 0x6d, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x1f, 0x0a, 0x0b, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x67, 0x72, 0x61, 0x69, 0x6e, 0x73, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x69, 0x6d, 0x65, 0x47, 0x72, 0x61, 0x69, 0x6e, 0x73,
	0x12, 0x46, 0x0a, 0x0a, 0x64, 0x69, 0x6d, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x05,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74,
	0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x56, 0x69,
	0x65, 0x77, 0x2e, 0x44, 0x69, 0x6d, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x64, 0x69,
	0x6d, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x40, 0x0a, 0x08, 0x6d, 0x65, 0x61, 0x73,
	0x75, 0x72, 0x65, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x72, 0x69, 0x6c,
	0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x73, 0x56, 0x69, 0x65, 0x77, 0x2e, 0x4d, 0x65, 0x61, 0x73, 0x75, 0x72, 0x65,
	0x52, 0x08, 0x6d, 0x65, 0x61, 0x73, 0x75, 0x72, 0x65, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x61,
	0x62, 0x65, 0x6c, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c,
	0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x1a, 0x57, 0x0a, 0x09, 0x44, 0x69, 0x6d, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x8d, 0x01, 0x0a, 0x07,
	0x4d, 0x65, 0x61, 0x73, 0x75, 0x72, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6c,
	0x61, 0x62, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x61, 0x62, 0x65,
	0x6c, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2a, 0x8d, 0x01, 0x0a, 0x0a,
	0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1b, 0x0a, 0x17, 0x4f, 0x42,
	0x4a, 0x45, 0x43, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43,
	0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x4f, 0x42, 0x4a, 0x45, 0x43,
	0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x54, 0x41, 0x42, 0x4c, 0x45, 0x10, 0x01, 0x12, 0x16,
	0x0a, 0x12, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x53, 0x4f,
	0x55, 0x52, 0x43, 0x45, 0x10, 0x02, 0x12, 0x15, 0x0a, 0x11, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54,
	0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4d, 0x4f, 0x44, 0x45, 0x4c, 0x10, 0x03, 0x12, 0x1c, 0x0a,
	0x18, 0x4f, 0x42, 0x4a, 0x45, 0x43, 0x54, 0x5f, 0x54, 0x59, 0x50, 0x45, 0x5f, 0x4d, 0x45, 0x54,
	0x52, 0x49, 0x43, 0x53, 0x5f, 0x56, 0x49, 0x45, 0x57, 0x10, 0x04, 0x42, 0xb5, 0x01, 0x0a, 0x13,
	0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x2e, 0x76, 0x31, 0x42, 0x0c, 0x43, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x72, 0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x69,
	0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x72, 0x75,
	0x6e, 0x74, 0x69, 0x6d, 0x65, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52, 0x52, 0x58, 0xaa, 0x02, 0x0f,
	0x52, 0x69, 0x6c, 0x6c, 0x2e, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x56, 0x31, 0xca,
	0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x5c, 0x56,
	0x31, 0xe2, 0x02, 0x1b, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea,
	0x02, 0x11, 0x52, 0x69, 0x6c, 0x6c, 0x3a, 0x3a, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x3a,
	0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rill_runtime_v1_catalog_proto_rawDescOnce sync.Once
	file_rill_runtime_v1_catalog_proto_rawDescData = file_rill_runtime_v1_catalog_proto_rawDesc
)

func file_rill_runtime_v1_catalog_proto_rawDescGZIP() []byte {
	file_rill_runtime_v1_catalog_proto_rawDescOnce.Do(func() {
		file_rill_runtime_v1_catalog_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_runtime_v1_catalog_proto_rawDescData)
	})
	return file_rill_runtime_v1_catalog_proto_rawDescData
}

var file_rill_runtime_v1_catalog_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_rill_runtime_v1_catalog_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_rill_runtime_v1_catalog_proto_goTypes = []interface{}{
	(ObjectType)(0),              // 0: rill.runtime.v1.ObjectType
	(Model_Dialect)(0),           // 1: rill.runtime.v1.Model.Dialect
	(*Table)(nil),                // 2: rill.runtime.v1.Table
	(*Source)(nil),               // 3: rill.runtime.v1.Source
	(*Model)(nil),                // 4: rill.runtime.v1.Model
	(*MetricsView)(nil),          // 5: rill.runtime.v1.MetricsView
	(*Source_ExtractPolicy)(nil), // 6: rill.runtime.v1.Source.ExtractPolicy
	(*Source_ExtractPolicy_ExtractConfig)(nil), // 7: rill.runtime.v1.Source.ExtractPolicy.ExtractConfig
	(*MetricsView_Dimension)(nil),              // 8: rill.runtime.v1.MetricsView.Dimension
	(*MetricsView_Measure)(nil),                // 9: rill.runtime.v1.MetricsView.Measure
	(*StructType)(nil),                         // 10: rill.runtime.v1.StructType
	(*structpb.Struct)(nil),                    // 11: google.protobuf.Struct
}
var file_rill_runtime_v1_catalog_proto_depIdxs = []int32{
	10, // 0: rill.runtime.v1.Table.schema:type_name -> rill.runtime.v1.StructType
	11, // 1: rill.runtime.v1.Source.properties:type_name -> google.protobuf.Struct
	10, // 2: rill.runtime.v1.Source.schema:type_name -> rill.runtime.v1.StructType
	6,  // 3: rill.runtime.v1.Source.policy:type_name -> rill.runtime.v1.Source.ExtractPolicy
	1,  // 4: rill.runtime.v1.Model.dialect:type_name -> rill.runtime.v1.Model.Dialect
	10, // 5: rill.runtime.v1.Model.schema:type_name -> rill.runtime.v1.StructType
	8,  // 6: rill.runtime.v1.MetricsView.dimensions:type_name -> rill.runtime.v1.MetricsView.Dimension
	9,  // 7: rill.runtime.v1.MetricsView.measures:type_name -> rill.runtime.v1.MetricsView.Measure
	7,  // 8: rill.runtime.v1.Source.ExtractPolicy.row:type_name -> rill.runtime.v1.Source.ExtractPolicy.ExtractConfig
	7,  // 9: rill.runtime.v1.Source.ExtractPolicy.file:type_name -> rill.runtime.v1.Source.ExtractPolicy.ExtractConfig
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_rill_runtime_v1_catalog_proto_init() }
func file_rill_runtime_v1_catalog_proto_init() {
	if File_rill_runtime_v1_catalog_proto != nil {
		return
	}
	file_rill_runtime_v1_schema_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_rill_runtime_v1_catalog_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Table); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rill_runtime_v1_catalog_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Source); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rill_runtime_v1_catalog_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Model); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rill_runtime_v1_catalog_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetricsView); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rill_runtime_v1_catalog_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Source_ExtractPolicy); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rill_runtime_v1_catalog_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Source_ExtractPolicy_ExtractConfig); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rill_runtime_v1_catalog_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetricsView_Dimension); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_rill_runtime_v1_catalog_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetricsView_Measure); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rill_runtime_v1_catalog_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rill_runtime_v1_catalog_proto_goTypes,
		DependencyIndexes: file_rill_runtime_v1_catalog_proto_depIdxs,
		EnumInfos:         file_rill_runtime_v1_catalog_proto_enumTypes,
		MessageInfos:      file_rill_runtime_v1_catalog_proto_msgTypes,
	}.Build()
	File_rill_runtime_v1_catalog_proto = out.File
	file_rill_runtime_v1_catalog_proto_rawDesc = nil
	file_rill_runtime_v1_catalog_proto_goTypes = nil
	file_rill_runtime_v1_catalog_proto_depIdxs = nil
}
