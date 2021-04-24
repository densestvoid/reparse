package structexp

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Bool struct {
	StructExp `structexp:"{{test}}"`
	Value     bool `structexp.name:"test"`
}

type Int struct {
	StructExp `structexp:"{{test}}"`
	Value     int `structexp.name:"test"`
}

type String struct {
	StructExp `structexp:"{{test}}"`
	Value     string `structexp.name:"test"`
}

type ParsableBool bool

func (p *ParsableBool) Parse(s string) error {
	if p == nil {
		return nil
	}

	if s == "a" {
		*p = true
	}

	return nil
}

type ParsableStruct struct {
	StructExp `structexp:"{{test}}"`
	Value     ParsableBool `structexp.name:"test" structexp.exp:"a|b"`
}

type NestedStruct struct {
	Value string `structexp.name:"test"`
}

type ParentNestedStruct struct {
	StructExp `structexp:"{{test}}"`
	Nested    NestedStruct
}

type EmbeddedStruct struct {
	Value string `structexp.name:"test"`
}

type ParentEmbeddedStruct struct {
	StructExp `structexp:"{{test}}"`
	EmbeddedStruct
}

type MissingFieldStruct struct {
	Value string `structexp.name:"test"`
}

func TestParse(t *testing.T) {
	type TestCase struct {
		Name     string
		String   string
		Input    interface{}
		Expected interface{}
		Error    error
	}

	testCases := []TestCase{
		{
			Name:     "Bool",
			String:   "true",
			Input:    &Bool{},
			Expected: &Bool{Value: true},
			Error:    nil,
		},
		{
			Name:     "Int",
			String:   "100",
			Input:    &Int{},
			Expected: &Int{Value: 100},
			Error:    nil,
		},
		{
			Name:     "String",
			String:   "string",
			Input:    &String{},
			Expected: &String{Value: "string"},
			Error:    nil,
		},
		{
			Name:     "ParsableField",
			String:   "a",
			Input:    &ParsableStruct{},
			Expected: &ParsableStruct{Value: ParsableBool(true)},
			Error:    nil,
		},
		{
			Name:     "NestedStruct",
			String:   "string",
			Input:    &ParentNestedStruct{Nested: NestedStruct{}},
			Expected: &ParentNestedStruct{Nested: NestedStruct{"string"}},
			Error:    nil,
		},
		{
			Name:     "EmbeddedStruct",
			String:   "string",
			Input:    &ParentEmbeddedStruct{EmbeddedStruct: EmbeddedStruct{}},
			Expected: &ParentEmbeddedStruct{EmbeddedStruct: EmbeddedStruct{"string"}},
			Error:    nil,
		},
		{
			Name:     "BoolNotStructError",
			String:   "true",
			Input:    false,
			Expected: false,
			Error:    &NotStruct{reflect.Bool},
		},
		{
			Name:     "IntNotStructError",
			String:   "",
			Input:    0,
			Expected: 0,
			Error:    &NotStruct{reflect.Int},
		},
		{
			Name:     "StringNotStructError",
			String:   "string",
			Input:    "",
			Expected: "",
			Error:    &NotStruct{reflect.String},
		},
		{
			Name:     "MissingFieldError",
			String:   "string",
			Input:    &MissingFieldStruct{},
			Expected: &MissingFieldStruct{},
			Error:    &MissingField{},
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			err := Parse(tc.String, tc.Input)
			assert.EqualValues(t, tc.Expected, tc.Input)
			assert.EqualValues(t, tc.Error, err)
		})
	}
}

func TestSetField(t *testing.T) {
	type TestCase struct {
		Name     string
		String   string
		Input    interface{}
		Expected interface{}
	}

	testCases := []TestCase{
		{
			Name:   "Bool",
			String: "true",
			Input:  new(bool),
			Expected: func() *bool {
				var b = true
				return &b
			}(),
		},
		{
			Name:   "Int",
			String: "100",
			Input:  new(int),
			Expected: func() *int {
				var i = 100
				return &i
			}(),
		},
		{
			Name:   "String",
			String: "string",
			Input:  new(string),
			Expected: func() *string {
				var s = "string"
				return &s
			}(),
		},
		{
			Name:   "ParsableField",
			String: "a",
			Input:  new(ParsableBool),
			Expected: func() *ParsableBool {
				var b = ParsableBool(true)
				return &b
			}(),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			require.NoError(t, setField(reflect.ValueOf(tc.Input), tc.String))
			assert.EqualValues(t, tc.Expected, tc.Input)
		})
	}
}

func TestUnderlyingValue(t *testing.T) {
	type TestCase struct {
		Name     string
		Value    reflect.Value
		Expected reflect.Value
	}

	testCases := []TestCase{
		{
			Name:     "Int",
			Value:    reflect.ValueOf(0),
			Expected: reflect.ValueOf(0),
		},
		{
			Name:     "Pointer",
			Value:    reflect.ValueOf(new(int)),
			Expected: reflect.ValueOf(0),
		},
		{
			Name:     "Interface",
			Value:    reflect.ValueOf(interface{}(0)),
			Expected: reflect.ValueOf(0),
		},
		{
			Name:     "InterfaceOfPointer",
			Value:    reflect.ValueOf(interface{}(new(int))),
			Expected: reflect.ValueOf(0),
		},
		{
			Name: "PointerToInterface",
			Value: reflect.ValueOf(func() *interface{} {
				var i interface{} = 0
				return &i
			}()),
			Expected: reflect.ValueOf(0),
		},
		{
			Name: "PointerToInterfaceOfPointerToInterface",
			Value: reflect.ValueOf(func() *interface{} {
				// reassignment of &i to i causes pointer loop,
				// so we create a new variable j
				var i interface{} = 0
				var j interface{} = &i
				return &j
			}()),
			Expected: reflect.ValueOf(0),
		},
	}

	for _, testCase := range testCases {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			assert.EqualValues(
				t,
				tc.Expected.Kind(),
				underlyingValue(tc.Value).Kind(),
			)
		})
	}
}
