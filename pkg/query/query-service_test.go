package query

import (
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestParseQuery(t *testing.T) {

	var tests = []struct {
		queryString string
		defaultDb   string
		want1       *mongoQuery
		error       string
	}{

		//Invalid Query
		{"wibble", "db1",
			nil, "'wibble' is not a valid MongoDB query expression"},

		//Invalid find arguments
		{"db.test.find(wibble)", "db1",
			nil, "the args 'wibble' to the 'find' method are not a valid"},

		//Invalid sort arguments
		{"db.test.find().sort(wibble)", "db1",
			nil, "the args 'wibble' to the 'sort' method are not a valid"},

		//Minimal find
		{"db.test.find()", "db1",
			&mongoQuery{"db1", "test", "find", primitive.D{}, nil, nil}, ""},

		//Minimal find with sort
		{"db.test.find().sort()", "db1",
			&mongoQuery{"db1", "test", "find", primitive.D{}, nil, nil}, ""},

		//Complex find
		{`db.test.find({"a": 10},{"_id": 0}).sort({"b": 1})`, "db1",
			&mongoQuery{"db1", "test", "find", primitive.D{{"a", int32(10)}}, primitive.D{{"_id", int32(0)}}, primitive.D{{"b", int32(1)}}}, ""},

		//Non default db
		{"db2.test.find()", "db1",
			&mongoQuery{"db2", "test", "find", primitive.D{}, nil, nil}, ""},

		//Minimal aggregate
		{"db.test.aggregate()", "db1",
			&mongoQuery{"db1", "test", "aggregate", primitive.A{}, nil, nil}, ""},
	}

	for _, test := range tests {
		if got1, err := parseQuery(test.queryString, test.defaultDb); test.want1 != nil && !reflect.DeepEqual(*got1, *test.want1) || err != nil && err.Error() != test.error {
			t.Errorf("parseQuery(%q, %q) = (%v,%v)", test.queryString, test.defaultDb, got1, err)
		}
	}
}

func TestTokenizeQuery(t *testing.T) {

	var tests = []struct {
		value string
		want1 []string
		want2 bool
	}{
		//Invalid: Empty string
		{"",
			nil, false},

		//Invalid: Missing collection
		{"db.find()",
			nil, false},

		//Invalid: wrong method name
		{"db.test.insert()",
			nil, false},

		//Invalid: Semi-colon
		{"db.test.find();",
			nil, false},

		//Valid: minimal find
		{"db.test.find()",
			[]string{"db", "test", "find", "", ""}, true},

		//Valid: explicit args find
		{"db.test.find({},{})", []string{"db", "test", "find", "{},{}", ""}, true},

		//Valid: explicit args sort
		{"db.test.find().sort({})", []string{"db", "test", "find", "", "{}"}, true},

		//Valid: explict args find & sort
		{"db.test.find({},{}).sort({})", []string{"db", "test", "find", "{},{}", "{}"}, true},

		//Valid: minimal aggregate
		{"db.test.aggregate()", []string{"db", "test", "aggregate", "", ""}, true},

		//Valid: explicit args aggregate
		{"db.test.aggregate([])", []string{"db", "test", "aggregate", "[]", ""}, true},

		//Valid: explicit args aggregate & sort
		{"db.test.aggregate([]).sort({})", []string{"db", "test", "aggregate", "[]", "{}"}, true},
	}

	for _, test := range tests {
		if got1, got2 := tokenizeQuery(test.value); !reflect.DeepEqual(got1, test.want1) || got2 != test.want2 {
			t.Errorf("tokenizeQuery(%q) = (%v,%v)", test.value, got1, got2)
		}
	}
}

func TestFindSentinel(t *testing.T) {

	var tests = []struct {
		value     string
		sentinels []string
		want1     int
		want2     int
	}{
		//Not present
		{"db.collection.find({})", []string{"aggregate("},
			-1, -1},

		//Present
		{"db.collection.find({})", []string{"find("},
			14, 19},
	}

	for _, test := range tests {
		if got1, got2 := findSentinel(test.value, test.sentinels...); got1 != test.want1 || got2 != test.want2 {
			t.Errorf("findSentinel(%q, %q) = (%v,%v)", test.value, test.sentinels, got1, got2)
		}
	}
}

func TestDecodeArgs(t *testing.T) {

	var tests = []struct {
		input string
		want1 interface{}
		want2 interface{}
		err   string
	}{
		//Invalid JSON
		{"fish",
			nil, nil, "invalid JSON literal."},

		//Minimal single argument
		{`{}`,
			primitive.D{}, nil, ""},

		//Minimal multiple arguments
		{`{}, {}`,
			primitive.D{}, primitive.D{}, ""},

		//Complex multiple Arguments
		{`{"a": true}, {"a": 1}`,
			primitive.D{{"a", true}}, primitive.D{{"a", int32(1)}}, ""},
	}

	for _, test := range tests {
		if got1, got2, err := decodeArgs(test.input); !reflect.DeepEqual(got1, test.want1) || !reflect.DeepEqual(got2, test.want2) || (err != nil && !strings.Contains(err.Error(), test.err)) {
			t.Errorf("decodeArgs(%q) = (%v,%v,%v)", test.input, got1, got2, err)
		}
	}
}

func TestPatchLiterals(t *testing.T) {

	var tests = []struct {
		input string
		want  interface{}
	}{
		//An array of Date fields
		{"[\"${new Date(5)}\"]",
			primitive.A{primitive.DateTime(5)}},

		//An array of ojects with a Date field.
		{"[{\"time\": \"${new Date(5)}\"}]",
			primitive.A{primitive.D{{"time", primitive.DateTime(5)}}}},

		//An object with a Date field.
		{"{\"time\": \"${new Date(5)}\"}",
			primitive.D{{"time", primitive.DateTime(5)}}},

		//An object with an array of Date fields.
		{"{\"times\": [\"${new Date(5)}\"]}",
			primitive.D{{"times", primitive.A{primitive.DateTime(5)}}}},
	}

	for _, test := range tests {

		var got interface{}
		err := bson.UnmarshalExtJSON([]byte(test.input), true, &got)
		if err == nil {
			patchLiterals(&got)
		}

		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("patchLiterals(%s) = %v ", got, test.want)
		}
	}

}

func TestConvertLiteral(t *testing.T) {

	var tests = []struct {
		input string
		want  interface{}
	}{
		{"", ""},
		{"${new Date(5)}", primitive.DateTime(5)},
	}

	for _, test := range tests {
		if got := convertLiteral(test.input); got != test.want {
			t.Errorf("convertLiteral(%q) = %q", test.input, got)
		}
	}
}
