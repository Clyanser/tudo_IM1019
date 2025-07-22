package maps

import (
	"fmt"
	"reflect"
)

func MapToStruct(data map[string]interface{}, dst interface{}) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr || dstVal.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dst must be a pointer to a struct")
	}

	dstVal = dstVal.Elem()
	dstType := dstVal.Type()

	for i := 0; i < dstVal.NumField(); i++ {
		fieldType := dstType.Field(i)
		fieldValue := dstVal.Field(i)

		// 获取 json tag
		tag := fieldType.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// 从 map 中查找对应字段
		val, ok := data[tag]
		if !ok {
			continue
		}

		// 处理字段赋值
		if err := setFieldValue(fieldValue, fieldType.Type, val); err != nil {
			return fmt.Errorf("error setting field %s: %v", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值，支持指针和基础类型
func setFieldValue(field reflect.Value, fieldType reflect.Type, val interface{}) error {
	valType := reflect.TypeOf(val)
	if valType == nil {
		return fmt.Errorf("value is nil")
	}

	// 如果字段是指针类型，则初始化指针指向的对象
	if fieldType.Kind() == reflect.Ptr {
		elemType := fieldType.Elem()
		elem := reflect.New(elemType).Elem()

		if err := setFieldValue(elem, elemType, val); err != nil {
			return err
		}

		field.Set(elem.Addr())
		return nil
	}

	// 获取值的反射
	valValue := reflect.ValueOf(val)

	// 如果字段类型和值类型一致，直接赋值
	if valValue.Type().AssignableTo(fieldType) {
		field.Set(valValue)
		return nil
	}

	// 否则尝试转换
	if valValue.Type().ConvertibleTo(fieldType) {
		field.Set(valValue.Convert(fieldType))
		return nil
	}

	return fmt.Errorf("cannot assign %v to %v", valValue.Type(), fieldType)
}
