package helper

import (
	"reflect"
	"strconv"
)

//  by chenyn
//将from结构体中的字段赋值到to结构体，字段之间可以通过keyMap映射
//keyMap的key为to的字段名，value为from的字段名
func LoadStruct(to interface{}, from interface{}, keyMap ...map[string]string) bool{
	typ1 := reflect.TypeOf(to)
	val1 := reflect.Indirect(reflect.ValueOf(to))
	if val1.Kind() != reflect.Struct || typ1.Kind() != reflect.Ptr{
		return false
	}
	typ1 = typ1.Elem()

	typ2 := reflect.TypeOf(from)
	if typ2.Kind() == reflect.Ptr {
		typ2 = typ2.Elem()
	}
	val2 := reflect.Indirect(reflect.ValueOf(from))
	if val2.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < val1.NumField(); i++ {
		if !val1.Field(i).CanSet() {
			continue
		}
		key := typ1.Field(i).Name
		if len(keyMap) > 0 {
			if _, ok := keyMap[0][typ1.Field(i).Name]; ok {
				key = keyMap[0][typ1.Field(i).Name]
			}
		}

		field,exists := typ2.FieldByName(key)
		if exists {
			rvf2 := val2.FieldByIndex(field.Index)
			if typ1.Field(i).Type.Kind() == field.Type.Kind() {
				val1.Field(i).Set(rvf2)
			} else {
				switch {
				case IsInt(val1.Field(i).Type()) && rvf2.Kind() == reflect.String:
					iv,e := strconv.Atoi(rvf2.String())
					if e == nil {
						val1.Field(i).SetInt(int64(iv))
					}
				case val1.Field(i).Kind() == reflect.String && IsInt(rvf2.Type()):
					val1.Field(i).SetString(strconv.Itoa(int(rvf2.Int())))
				}
			}
		}
	}
	return true
}
