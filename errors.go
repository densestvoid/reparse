package structexp // nolint:golint // in another file

import (
	"fmt"
	"reflect"
)

// InvalidType occurs when trying to set the value of an unaddreesable type
type InvalidType struct {
	reflect.Type
}

func (err InvalidType) Error() string {
	return "value of type %T unable to be set (must be addressable)"
}

// NotStruct occurs when anything but a pointer to a struct is passed into Parse
type NotStruct struct {
	K reflect.Kind
}

func (err *NotStruct) Error() string {
	return fmt.Sprintf(
		"object to parse is not %v, is %v",
		reflect.Struct,
		err.K,
	)
}

// MissingField occurs when the struct to be parsed does not have a StructExp field
type MissingField struct{}

func (err *MissingField) Error() string {
	return fmt.Sprintf("object missing field with type %T", StructExp{})
}

// NoMatch occurs when the string to be parsed does not matc hthe built regular expression
type NoMatch struct{}

func (err *NoMatch) Error() string {
	return "object regular expression has no matches for the input"
}
