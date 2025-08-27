package redigo

import "reflect"

func isBasicType(v interface{}) bool {
	if v == nil {
		return true
	}

	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice:
		// 检查是否是[]byte
		return t.Elem().Kind() == reflect.Uint8
	default:
		return false
	}
}
