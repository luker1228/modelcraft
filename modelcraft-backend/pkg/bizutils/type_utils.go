package bizutils

import "reflect"

// IsArrayType 检查给定的值是否为数组或切片类型。
// 返回 true 如果值的类型是数组或切片，否则返回 false。
func IsArrayType(i any) bool {
	v := reflect.ValueOf(i)
	return v.Kind() == reflect.Array || v.Kind() == reflect.Slice
}
