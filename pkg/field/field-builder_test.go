package field

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestFieldBuild(t *testing.T) {

	strValue := "value"
	var int64Value int64 = 20
	strInt64Value := strconv.FormatInt(int64Value, 10)
	fieldName := "test"

	var tests = []struct {
		values                []interface{}
		wantAllValuesSameType bool
		wantFieldType         data.FieldType
		wantFirstValue        interface{}
		wantBuild             *data.Field
	}{
		//empty
		{[]interface{}{}, true,
			data.FieldTypeString, nil, data.NewField(fieldName, nil, []string{})},

		//nil value
		{[]interface{}{nil}, true,
			data.FieldTypeNullableString, nil, data.NewField(fieldName, nil, []*string{nil})},

		//string value
		{[]interface{}{strValue},
			true, data.FieldTypeString, strValue, data.NewField(fieldName, nil, []string{"value"})},

		//nil and string value
		{[]interface{}{nil, strValue},
			true, data.FieldTypeNullableString, strValue, data.NewField(fieldName, nil, []*string{nil, &strValue})},

		//string and nil value
		{[]interface{}{strValue, nil},
			true, data.FieldTypeNullableString, strValue, data.NewField(fieldName, nil, []*string{&strValue, nil})},

		//string and int64 value
		{[]interface{}{strValue, int64Value},
			false, data.FieldTypeString, strValue, data.NewField(fieldName, nil, []string{strValue, strInt64Value})},

		//int64 and string value
		{[]interface{}{int64Value, strValue},
			false, data.FieldTypeString, int64Value, data.NewField(fieldName, nil, []string{strInt64Value, strValue})},

		//string, int64 and nil value
		{[]interface{}{strValue, int64Value, nil},
			false, data.FieldTypeNullableString, strValue, data.NewField(fieldName, nil, []*string{&strValue, &strInt64Value, nil})},
	}

	for _, test := range tests {

		field := newField("test", 1)
		for _, v := range test.values {
			field.append(v)
		}

		if got1 := field.allValuesSameType(); got1 != test.wantAllValuesSameType {
			t.Errorf("%v field.allValuesSameType() = (%v)", test.values, got1)
		}

		if got2 := field.fieldType(); got2 != test.wantFieldType {
			t.Errorf("%v field.fieldType() = (%v)", test.values, got2)
		}

		if got3 := field.firstValue(); got3 != test.wantFirstValue {
			t.Errorf("%v field.firstValue() = (%v)", test.values, got3)
		}

		if got4 := field.build(); !reflect.DeepEqual(got4, test.wantBuild) {
			t.Errorf("%v field.build() = (%v)", test.values, *got4)
		}
	}
}

func TestFieldExpandTo(t *testing.T) {

	fieldName := "test"
	strValue := "value"

	var tests = []struct {
		values    []interface{}
		size      int
		wantBuild *data.Field
	}{
		//empty expand to -1
		{[]interface{}{}, -1,
			data.NewField(fieldName, nil, []string{})},

		//empty expand to 0
		{[]interface{}{}, 0,
			data.NewField(fieldName, nil, []string{})},

		//empty expand to 2
		{[]interface{}{}, 2,
			data.NewField(fieldName, nil, []*string{nil, nil})},

		//single value expand to 0
		{[]interface{}{strValue}, 0,
			data.NewField(fieldName, nil, []string{strValue})},

		//string value expand to 2
		{[]interface{}{strValue}, 2,
			data.NewField(fieldName, nil, []*string{&strValue, nil})},
	}

	for _, test := range tests {

		field := newField(fieldName, 1)
		for _, v := range test.values {
			field.append(v)
		}
		field.expandTo(test.size)

		if got := field.build(); !reflect.DeepEqual(got, test.wantBuild) {
			t.Errorf("%v field.build() = (%v)", test.values, *got)
		}
	}

}

func TestAsFieldValue(t *testing.T) {

	objectId := primitive.NewObjectID()
	decimal128, _ := primitive.ParseDecimal128("555")
	dbPointer := primitive.DBPointer{DB: "mydb", Pointer: objectId}
	now := time.Unix(1620586358172, 0)
	datetime := primitive.NewDateTimeFromTime(now)
	var flt32 float32 = 34.5
	var bigint int64 = 32
	var binData = primitive.Binary{Subtype: 0, Data: []byte("hello world")}

	var tests = []struct {
		value interface{}
		want1 interface{}
	}{
		//float 32
		{flt32, flt32},

		//string
		{"hello world", "hello world"},

		//Ordered object
		{primitive.D{{Key: "a", Value: 20}}, "{\"a\":20}"},

		//Unordered object
		{primitive.M{"b": 45}, "{\"b\":45}"},

		//Array
		{primitive.A{1, 5}, "[1,5]"},

		//BinData
		{binData, "BinData(0, \"aGVsbG8gd29ybGQ=\")"},

		//undefined
		{primitive.Undefined{}, "undefined"},

		//Object Id
		{objectId, fmt.Sprintf("ObjectId(%q)", objectId.Hex())},

		//boolean
		{true, true},

		//DateTime
		{datetime, now},

		//null
		{primitive.Null{}, "null"},

		//Regex
		{primitive.Regex{Pattern: ".+", Options: "g"}, "/.+/g"},

		//DBPointer
		{dbPointer,
			fmt.Sprintf("DBPointer(%q, ObjectId(%q))", "mydb", objectId.Hex())},
		//integer
		{32, 32},
		{bigint, bigint},
		{decimal128, "555"},
	}

	for _, test := range tests {
		if got1 := asFieldValue(test.value); !reflect.DeepEqual(got1, test.want1) {
			t.Errorf("asFieldValue(%v) = (%v)", test.value, got1)
		}
	}
}

func TestFieldBuilderBuild(t *testing.T) {

	strColName := "strCol"
	strValue := "Tom"
	intColName := "intCol"
	intValue1 := int32(52)
	intValue2 := int32(22)

	var tests = []struct {
		records []primitive.D
		want    []*data.Field
	}{
		//Should have 2 columns with different data types.
		{
			[]primitive.D{
				primitive.D{{strColName, strValue}, {intColName, intValue1}},
			},
			[]*data.Field{
				data.NewField(strColName, nil, []string{strValue}),
				data.NewField(intColName, nil, []int32{intValue1}),
			},
		},

		//Should pad column with missing value with nil
		{
			[]primitive.D{
				primitive.D{{strColName, strValue}, {intColName, intValue1}},
				primitive.D{{intColName, intValue2}},
			},
			[]*data.Field{
				data.NewField(strColName, nil, []*string{&strValue, nil}),
				data.NewField(intColName, nil, []int32{intValue1, intValue2}),
			},
		},

		//Should pad column with missing value with nil
		{
			[]primitive.D{
				primitive.D{{intColName, intValue1}},
				primitive.D{{strColName, strValue}, {intColName, intValue2}},
			},
			[]*data.Field{
				data.NewField(intColName, nil, []int32{intValue1, intValue2}),
				data.NewField(strColName, nil, []*string{nil, &strValue}),
			},
		},

		//Should have column with mixed values converted to string.
		{
			[]primitive.D{
				primitive.D{{intColName, intValue1}},
				primitive.D{{intColName, strValue}},
			},
			[]*data.Field{
				data.NewField(intColName, nil, []string{"52", strValue}),
			},
		},
	}

	for _, test := range tests {
		fieldBuilder := NewFieldBuilder(5)
		for _, record := range test.records {
			fieldBuilder.ProcessRecord(record)
		}

		got := fieldBuilder.BuildFields()
		for i, got1 := range got {
			want1 := test.want[i]
			if !reflect.DeepEqual(*got1, *want1) {
				t.Errorf("field[%d] %v == %v", i, *want1, *got1)
			}
		}
	}

}
