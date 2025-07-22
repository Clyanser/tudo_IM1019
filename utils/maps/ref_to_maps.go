package maps

import "reflect"

// 编写反射函数
func RefToMap(data any, tag string) map[string]any {
	maps := map[string]any{}
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		getTag, ok := field.Tag.Lookup(tag)
		if !ok {
			continue
		}
		val := v.Field(i)
		if val.IsZero() {
			continue
		}
		if field.Type.Kind() == reflect.Struct {
			newMaps := RefToMap(val.Elem().Interface(), tag)
			maps[getTag] = newMaps
			continue
		}
		if field.Type.Kind() == reflect.Ptr {
			if field.Type.Elem().Kind() == reflect.Struct {
				newMaps := RefToMap(val.Elem().Interface(), tag)
				maps[getTag] = newMaps
				continue
			}
			maps[getTag] = val.Elem().Interface()
			continue
		}
		maps[getTag] = val.Interface()

	}
	return maps
}
