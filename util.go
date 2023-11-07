package gomock

import (
	"math"
	"reflect"
)

func sumSlice[T int64](values []T) T {
	var sum T = 0
	for _, value := range values {
		sum += value
	}
	return sum
}

// merge gt, gte
func makeGteVal[T int | int64](kind reflect.Kind, gt T, gte T,
	gtExists, gteExists bool) (T, bool) {
	if !gtExists && !gteExists {
		return T(intMinVal(kind)), false
	}
	if !gtExists {
		return gte, true
	}
	if !gteExists {
		return gt + 1, true
	}
	if gt+1 < gte {
		return gt + 1, true
	}
	return gte, true
}

// merge lt, lte
func makeLtVal[T int | int64](kind reflect.Kind, lt, lte T,
	ltExists, lteExists bool) (T, bool) {
	if !ltExists && !lteExists {
		return T(intMaxVal(kind)), false
	}
	if !ltExists {
		return lte + 1, true
	}
	if !lteExists {
		return lt, true
	}
	if lte+1 > lt {
		return lte + 1, true
	}
	return lt, true
}
func intMinVal(kind reflect.Kind) int64 {
	switch kind {
	case reflect.Int:
		return math.MinInt
	case reflect.Int8:
		return math.MinInt8
	case reflect.Int16:
		return math.MinInt16
	case reflect.Int32:
		return math.MinInt32
	case reflect.Int64:
		return math.MinInt64
	case reflect.Uint:
		return 0
	case reflect.Uint8:
		return 0
	case reflect.Uint16:
		return 0
	case reflect.Uint32:
		return 0
	case reflect.Uint64:
		return 0
	}
	return 0
}
func intMaxVal(kind reflect.Kind) int64 {
	switch kind {
	case reflect.Int, reflect.Uint:
		return math.MaxInt
	case reflect.Int8, reflect.Uint8:
		return math.MaxInt8
	case reflect.Int16, reflect.Uint16:
		return math.MaxInt16
	case reflect.Int32, reflect.Uint32:
		return math.MaxInt32
	case reflect.Int64, reflect.Uint64:
		return math.MaxInt64
	}
	return 0
}

func maxFunc[T int | int64 | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func mapToSlice[T string | struct{}](valueMap map[T]T, fn func(T, T) T) []T {
	values := make([]T, 0, len(valueMap))
	for key, value := range valueMap {
		values = append(values, fn(key, value))
	}
	return values
}
