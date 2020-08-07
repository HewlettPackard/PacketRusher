package tlv

import (
	"reflect"
)

func isNumber(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

func isRefType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer,
		reflect.Interface, reflect.Slice:
		return true
	default:
		return false
	}
}

func hasValue(value reflect.Value) bool {
	if isRefType(value.Type()) {
		return !value.IsNil()
	} else {
		return value.IsValid()
	}
}
