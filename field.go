package structexp // nolint:golint // in another file

import (
	"fmt"
	"reflect"
)

const (
	captureGroupNameKey = "structexp.name"
	expKey              = "structexp.exp"
)

// Default regular expression used when parsing struct fields
const (
	DefaultBoolRegexp   = `1|t|T|TRUE|true|True|0|f|F|FALSE|false|False`
	DefaultIntRegexp    = `[[:digit:]]+`
	DefaultStringRegexp = `[[:print:]]+`
)

func kindExp(k reflect.Kind) string {
	// nolint:exhaustive // unnecessary
	switch k {
	case reflect.Bool:
		return DefaultBoolRegexp
	case reflect.Int:
		return DefaultIntRegexp
	case reflect.String:
		return DefaultStringRegexp
	default:
		return ""
	}
}

type field struct {
	Value            reflect.Value
	CaptureGroupName string
	Exp              string
}

func newField(value reflect.Value, reflectField *reflect.StructField) *field {
	f := &field{
		Value:            value,
		CaptureGroupName: reflectField.Name,
		Exp:              kindExp(reflectField.Type.Kind()),
	}

	if captureGroupName := reflectField.Tag.Get(captureGroupNameKey); captureGroupName != "" {
		f.CaptureGroupName = captureGroupName
	}

	if exp := reflectField.Tag.Get(expKey); exp != "" {
		f.Exp = exp
	}

	return f
}

func (f field) NamedCaptureGroup() string {
	return fmt.Sprintf("(?P<%s>%s)", f.CaptureGroupName, f.Exp)
}
