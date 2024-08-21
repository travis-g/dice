package main

import "reflect"

func deepEqual(a, b interface{}) bool {
	return deepEqualValue(reflect.ValueOf(a), reflect.ValueOf(b))
}

func deepEqualValue(a, b reflect.Value) bool {
	if !a.IsValid() || !b.IsValid() {
		return a.IsValid() == b.IsValid()
	}

	if a.Type() != b.Type() {
		return false
	}

	switch a.Kind() {
	case reflect.Ptr:
		if a.IsNil() || b.IsNil() {
			return a.IsNil() == b.IsNil()
		}
		return deepEqualValue(a.Elem(), b.Elem())

	case reflect.Struct:
		for i := 0; i < a.NumField(); i++ {
			if !deepEqualValue(a.Field(i), b.Field(i)) {
				return false
			}
		}
		return true

	case reflect.Slice:
		if a.Len() != b.Len() {
			return false
		}
		for i := 0; i < a.Len(); i++ {
			if !deepEqualValue(a.Index(i), b.Index(i)) {
				return false
			}
		}
		return true

	case reflect.Map:
		if a.Len() != b.Len() {
			return false
		}
		for _, key := range a.MapKeys() {
			if !deepEqualValue(a.MapIndex(key), b.MapIndex(key)) {
				return false
			}
		}
		return true

	default:
		return reflect.DeepEqual(a.Interface(), b.Interface())
	}
}
