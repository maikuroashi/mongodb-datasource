package main

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type field struct {
	Name     string
	Nullable bool
	Values   []interface{}
}

func newField(name string, capacity int) *field {

	return &field{
		Name:   name,
		Values: make([]interface{}, 0, capacity),
	}
}

func (f *field) append(value interface{}) {
	f.Nullable = f.Nullable || value == nil
	f.Values = append(f.Values, asFieldValue(value))
}

func (f *field) exendTo(size int) {
	for len(f.Values) < size {
		f.append(nil)
	}
}

func (f *field) fieldType() data.FieldType {

	fieldType := data.FieldTypeString

	if f.allValuesSameType() {
		firstValue := f.firstValue()
		if firstValue != nil {
			fieldType = fieldTypeFromValue(firstValue)
		}
	}

	if f.Nullable {
		fieldType = fieldType.NullableType()
	}

	return fieldType
}

func (f *field) build() *data.Field {

	fieldType := f.fieldType()
	field := data.NewFieldFromFieldType(fieldType, len(f.Values))
	field.Name = f.Name
	for i, v := range f.Values {
		v = f.convertValue(fieldType, v)
		field.Set(i, v)
	}
	return field
}

func (f *field) allValuesSameType() bool {

	result := true
	var lastType reflect.Type = nil

	for _, v := range f.Values {

		if lastType == nil && v != nil {
			lastType = reflect.TypeOf(v)
		} else if v != nil && lastType != reflect.TypeOf(v) {
			result = false
			break
		}
	}
	return result
}

func (f *field) firstValue() interface{} {

	var result interface{}
	for _, v := range f.Values {
		if v != nil {
			result = v
			break
		}
	}
	return result
}

func (f *field) convertValue(fieldType data.FieldType, value interface{}) interface{} {

	result := value

	if fieldType.NullableType() == data.FieldTypeNullableString && result != nil {
		resultType := fieldTypeFromValue(result).NullableType()
		if resultType != data.FieldTypeNullableString {
			result = fmt.Sprintf("%v", result)
		}
	}

	if f.Nullable {
		result = asNullableValue(result)
	}
	return result
}

func asFieldValue(value interface{}) interface{} {

	switch value.(type) {

	case primitive.ObjectID:
		v := value.(primitive.ObjectID)
		return v.String()

	case primitive.Undefined:
		return "undefined"

	case primitive.DateTime:
		v := value.(primitive.DateTime)
		return v.Time()

	case primitive.Decimal128:
		v := value.(primitive.Decimal128)
		return v.String()

	case primitive.Regex:
		v := value.(primitive.Regex)
		return fmt.Sprintf("/%s/%s", v.Pattern, v.Options)

	case primitive.Binary:
		v := value.(primitive.Binary)
		return fmt.Sprintf(`BinData(%d, "%s")`, v.Subtype, base64.StdEncoding.EncodeToString(v.Data))

	case primitive.DBPointer:
		v := value.(primitive.DBPointer)
		return fmt.Sprintf(`DBPointer("%s", %s)`, v.DB, v.Pointer)

	case primitive.A:
		return asJsonString(value)

	case primitive.D:
		return asJsonString(value)

	default:
		return value
	}
}

func asJsonString(value interface{}) string {

	json, err := bson.MarshalExtJSON(value, true, false)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(json)
}

func asNullableValue(value interface{}) interface{} {
	switch value.(type) {

	case int8:
		v := value.(int8)
		return &v
	case int16:
		v := value.(int16)
		return &v
	case int32:
		v := value.(int32)
		return &v
	case int64:
		v := value.(int64)
		return &v
	case uint8:
		v := value.(uint8)
		return &v
	case uint16:
		v := value.(uint16)
		return &v
	case uint32:
		v := value.(uint32)
		return &v
	case uint64:
		v := value.(uint64)
		return &v
	case float32:
		v := value.(float32)
		return &v
	case float64:
		v := value.(float64)
		return &v
	case string:
		v := value.(string)
		return &v
	case bool:
		v := value.(bool)
		return &v
	case time.Time:
		v := value.(time.Time)
		return &v
	default:
		return value
	}
}

func fieldTypeFromValue(value interface{}) data.FieldType {
	switch value.(type) {
	// ints
	case int8:
		return data.FieldTypeInt8
	case *int8:
		return data.FieldTypeNullableInt8
	case int16:
		return data.FieldTypeInt16
	case *int16:
		return data.FieldTypeNullableInt16
	case int32:
		return data.FieldTypeInt32
	case *int32:
		return data.FieldTypeNullableInt32
	case int64:
		return data.FieldTypeInt64
	case *int64:
		return data.FieldTypeNullableInt64

	// uints
	case uint8:
		return data.FieldTypeUint8
	case *uint8:
		return data.FieldTypeNullableUint8
	case uint16:
		return data.FieldTypeUint16
	case *uint16:
		return data.FieldTypeNullableUint16
	case uint32:
		return data.FieldTypeUint32
	case *uint32:
		return data.FieldTypeNullableUint32
	case uint64:
		return data.FieldTypeUint64
	case *uint64:
		return data.FieldTypeNullableUint64

	// floats
	case float32:
		return data.FieldTypeFloat32
	case *float32:
		return data.FieldTypeNullableFloat32
	case float64:
		return data.FieldTypeFloat64
	case *float64:
		return data.FieldTypeNullableFloat64

	// others
	case string:
		return data.FieldTypeString
	case *string:
		return data.FieldTypeNullableString
	case bool:
		return data.FieldTypeBool
	case *bool:
		return data.FieldTypeNullableBool
	case time.Time:
		return data.FieldTypeTime
	case *time.Time:
		return data.FieldTypeNullableTime

	default:
		panic("unsupported type")
	}
}

type FieldBuilder struct {
	recordCount int
	fields      []*field
	index       map[string]*field
}

func NewFieldBuilder(capacity int) *FieldBuilder {
	return &FieldBuilder{
		fields: make([]*field, 0, capacity),
		index:  make(map[string]*field),
	}
}

func (fb *FieldBuilder) ProcessRecord(record primitive.D) {

	for _, e := range record {
		field := fb.field(e.Key)
		field.exendTo(fb.recordCount)
		field.append(e.Value)
	}
	fb.recordCount++
}

func (fb *FieldBuilder) BuildFields() []*data.Field {

	fields := make([]*data.Field, 0, fb.recordCount)
	for _, field := range fb.fields {
		field.exendTo(fb.recordCount)
		fields = append(fields, field.build())
	}
	return fields
}

func (fb *FieldBuilder) field(name string) *field {

	result := fb.index[name]
	if result == nil {
		result = newField(name, 10)
		fb.fields = append(fb.fields, result)
		fb.index[name] = result
	}
	return result
}
