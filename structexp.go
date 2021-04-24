// Package structexp parses strings into structs using regular expressions
//
// Currently accepted struct field types:
//  - bool
//  - int
//  - string
//  - ParsableField
//
// Struct variable tags:
//  - structexp: used with the StructExp type to define the regular expression used for parsing
//  - structexp.name: the variable regexp capture group name and string wrapped in double curly
//    braces {{}} to replace in the regular expression
//  - structexp.exp: the variable regular expression to use in the named capture group
//
// Notes:
//  - bool values are parsed from the regexp string result using strconv.ParseBool.
//    This is why the DefaultBoolExp value is `1|t|T|TRUE|true|True|0|f|F|FALSE|false|False`
//  - int values are parsed from the regexp string result using strconv.ParseInt.
//    This is why the DefaultIntExp value is `[[:digit:]]+`
//  - It is not recommended to set the structexp.exp tag for bool or int fields,
//    as this will likely make them unable to be parsed. Instead, define a type that
//    satisfies the ParsableField interface
//  - ParsableFields need the structexp.exp tag set
//  - Nested and Embedded structs are supported
//
// Example:
//
//  // Evaluated regex would be:
//  // `^bool:(?P<B>1|t|T|TRUE|true|True|0|f|F|FALSE|false|False), int:(?P<integer>[[:digit:]]+), string:(?P<str>\d+\s+\W+), parsable:(?P<P>parse)`
//  type Example struct {
//      StructExp `structexp="^bool: {{B}}, int: {{integer}}, string: {{str}}, parsable: {{P}}"`
//      Bool bool `structexp.name="B"`
//      Int int `structexp.name="integer"`
//      String string `structexp.name="str" structexp.exp="\d+\s+\W+"`
//      Parsable ParsableField `structexp.name="P" structexp.exp="[pP]ars(abl)?e"`
//  }
//
package structexp

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const tagKey = "structexp"

// StructExp is a required field for a struct that will be parsed,
// to apply the structexp tag as the base regular expression
type StructExp struct{}

// ParsableField interface defines a means of converting the regex
// string result to a type other than a bool, int, or string; or,
// changing how one of those types should be parsed. For example
// if an int regex matches text with commas, a CommaInt type might be
// defined to remove the commas before parsing.
type ParsableField interface {
	Parse(string) error
}

// Parse uses the struct argument's fields to construct a regular
// expression with named capture groups to parse the struct fields
// from the string argument.
//
// Errors occur if:
//  - argument is not the address of a struct
//  - struct is missing a StructExp field
//  - regular expression does not match the string
func Parse(s string, i interface{}) error {
	// Verify interface is a pointer to a structure
	t := reflect.TypeOf(i)
	if kind := t.Kind(); kind != reflect.Ptr {
		return &NotStruct{kind}
	}

	t = t.Elem()
	if kind := t.Kind(); kind != reflect.Struct {
		return &NotStruct{kind}
	}

	base, err := regexpBase(t)
	if err != nil {
		return err
	}
	fields := listFields(reflect.ValueOf(i).Elem())
	regxp, err := fillRegexp(base, fields)
	if err != nil {
		return err
	}

	if !regxp.MatchString(s) {
		return &NoMatch{}
	}

	matches := regxp.FindStringSubmatch(s)
	for _, field := range fields {
		if idx := regxp.SubexpIndex(field.CaptureGroupName); idx != -1 {
			if err := setField(field.Value, matches[idx]); err != nil {
				return err
			}
		}
	}

	return nil
}

// Get the Regexp base from the Regexp field
func regexpBase(t reflect.Type) (string, error) {
	regexpField, ok := t.FieldByNameFunc(func(name string) bool {
		if field, _ := t.FieldByName(name); field.Type == reflect.TypeOf(StructExp{}) {
			return true
		}
		return false
	})
	if !ok {
		return "", &MissingField{}
	}
	return regexpField.Tag.Get(tagKey), nil
}

func listFields(v reflect.Value) []*field {
	t := v.Type()

	var fields []*field
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip the Regexp field
		if field.Type == reflect.TypeOf(StructExp{}) {
			continue
		}

		// nolint:exhaustive // unnecessary
		switch field.Type.Kind() {
		case reflect.Bool:
		case reflect.Int:
		case reflect.String:
		default:
			if reflect.PtrTo(field.Type).Implements(reflect.TypeOf((*ParsableField)(nil)).Elem()) {
				break
			}
			if field.Type.Kind() == reflect.Struct {
				fields = append(fields, listFields(v.Field(i))...)
			}
			continue
		}

		fields = append(fields, newField(v.Field(i), &field))
	}
	return fields
}

// Fill in the regexp string with field expressions
func fillRegexp(base string, fields []*field) (*regexp.Regexp, error) {
	for _, field := range fields {
		base = strings.Replace(
			base,
			fmt.Sprintf("{{%s}}", field.CaptureGroupName),
			field.NamedCaptureGroup(),
			1,
		)
	}
	return regexp.Compile(base)
}

func setField(val reflect.Value, s string) error {
	underVal := underlyingValue(val)

	// Underlying value must be settable
	if !underVal.CanSet() {
		return &InvalidType{val.Type()}
	}

	// Check if pointer to underlying type satisfies the ParsableFiled interface
	if underVal.CanAddr() {
		if parsable, ok := underVal.Addr().Interface().(ParsableField); ok {
			return parsable.Parse(s)
		}
	}

	// Set the fields of the basic types
	// nolint:exhaustive // unnecessary
	switch underVal.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		underVal.SetBool(b)
	case reflect.Int:
		i, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return err
		}
		underVal.SetInt(i)
	case reflect.String:
		underVal.SetString(s)
	}

	return nil
}

func underlyingValue(value reflect.Value) reflect.Value {
	for exit := false; !exit; {
		// nolint:exhaustive // unnecessary
		switch value.Kind() {
		case reflect.Interface, reflect.Ptr:
			value = value.Elem()
		default:
			exit = true
		}
	}
	return value
}
