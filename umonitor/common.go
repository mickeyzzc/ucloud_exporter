package umonitor

import (
	"errors"
	"strconv"
)

func interface2Float64(inter interface{}) (float64, error) {

	switch inter.(type) {

	case int:
		return float64(inter.(int)), nil
	case int32:
		return float64(inter.(int32)), nil
	case int64:
		return float64(inter.(int64)), nil
	case float32:
		return float64(inter.(float32)), nil
	case float64:
		return inter.(float64), nil
	}
	return 0.0, errors.New("not change")
}

func MergeSlice(s1 []interface{}, s2 []interface{}) []interface{} {
	slice := make([]interface{}, len(s1)+len(s2))
	copy(slice, s1)
	copy(slice[len(s1):], s2)
	return slice
}

func Interface2String(inter interface{}) string {

	switch inter.(type) {

	case string:
		return inter.(string)
	case int:
		return string(inter.(int))
	case float64:
		return strconv.FormatFloat(inter.(float64), 'g', -1, 64)
	}
	return ""
}
