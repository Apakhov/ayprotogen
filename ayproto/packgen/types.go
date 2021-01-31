package packgen

import "reflect"

type FieldType int

func nameFromReflect(k reflect.Kind) (string, error) {
	switch k {
	case reflect.Int:
		return "int", nil
	case reflect.Int8:
		return "int8", nil
	case reflect.Int16:
		return "int16", nil
	case reflect.Int32:
		return "int32", nil
	case reflect.Int64:
		return "int64", nil
	case reflect.Uint:
		return "uint", nil
	case reflect.Uint8:
		return "uint8", nil
	case reflect.Uint16:
		return "uint16", nil
	case reflect.Uint32:
		return "uint32", nil
	case reflect.Uint64:
		return "uint64", nil
	case reflect.String:
		return "string", nil
	default:
		return "", ErrBadType
	}
}
