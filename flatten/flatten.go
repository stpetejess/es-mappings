package flatten

import (
	"fmt"
	"reflect"
)

func Flatten(thing map[string]interface{}, doFlattenSlice bool) map[string]interface{} {
	result := make(map[string]interface{})

	for k, raw := range thing {
		flatten(result, k, reflect.ValueOf(raw), doFlattenSlice)
	}

	return result
}

func flatten(result map[string]interface{}, prefix string, v reflect.Value, doFlattenSlice bool) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		flattenMap(result, prefix, v, doFlattenSlice)
	case reflect.Slice:
		if doFlattenSlice {
			flattenSlice(result, prefix, v, doFlattenSlice)
		} else {
			result[prefix] = v.Interface()
		}
	default:
		result[prefix] = v.Interface()
	}
}

func flattenMap(result map[string]interface{}, prefix string, v reflect.Value, doFlattenSlice bool) {
	for _, k := range v.MapKeys() {
		if k.Kind() == reflect.Interface {
			k = k.Elem()
		}

		if k.Kind() != reflect.String {
			panic(fmt.Sprintf("%s: map key is not string: %s", prefix, k))
		}

		flatten(result, fmt.Sprintf("%s.%s", prefix, k.String()), v.MapIndex(k), doFlattenSlice)
	}
}

func flattenSlice(result map[string]interface{}, prefix string, v reflect.Value, doFlattenSlice bool) {
	prefix = prefix + "."
	for i := 0; i < v.Len(); i++ {
		flatten(result, fmt.Sprintf("%s%d", prefix, i), v.Index(i), doFlattenSlice)
	}
}
