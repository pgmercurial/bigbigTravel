package helper

import "reflect"

func IsInt(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func IsFloat(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Float64, reflect.Float32:
		return true
	default:
		return false
	}
}

func IsString(t reflect.Type) bool {
	return t.Kind() == reflect.String
}

func IsNumeric(t reflect.Type) bool {
	return IsInt(t) || IsFloat(t)
}

func IsScalar(t reflect.Type) bool {
	return IsInt(t) || IsFloat(t) || IsString(t)
}
