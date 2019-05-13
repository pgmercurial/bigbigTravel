package helper

import (
	"strconv"
	"strings"
	"fmt"
	"reflect"
	"bytes"
	"crypto/md5"
	"encoding/hex"
)

func SplitAsIntSlice(str string, sep string) []int {
	strs := strings.Split(str, sep)
	ints := make([]int,0,len(strs))
	for _,s := range strs{
		i, err := strconv.ParseInt(s, 10, 32)
		if err != nil{
			continue
		}
		ints = append(ints, int(i))
	}
	return ints
}

func JoinIntSlice(ints []int, sep string) string {
	strs := make([]string, 0, len(ints))
	for _,i := range ints {
		strs = append(strs, fmt.Sprint(i))
	}
	return strings.Join(strs, sep)
}

func JoinStringSlice(ss []string, sep string) string {
	strs := make([]string, 0, len(ss))
	for _,i := range ss {
		strs = append(strs, fmt.Sprint(i))
	}
	return strings.Join(strs, sep)
}

func StrUnderline2CamelCase(str string, firstUpper bool) string {
	strSlice := strings.Split(str, "_")
	for i,str := range strSlice {
		bts := []byte(str)
		if i > 0 || firstUpper {
			bts[0] = bytes.ToUpper(bts[:1])[0]
		}
		strSlice[i] = string(bts)
	}
	return strings.Join(strSlice, "")
}

func Ucfirst(str string) string {
	bts := []byte(str)
	bts[0] = bytes.ToUpper(bts[:1])[0]
	return string(bts)
}

func Lcfirst(str string) string {
	bts := []byte(str)
	bts[0] = bytes.ToLower(bts[:1])[0]
	return string(bts)
}

func Md5(string string) string {
	b := []byte(string)
	m := md5.Sum(b)
	return hex.EncodeToString(m[:])
}


func ToSting(vs ...interface{}) string {
	result := ""
	for n,i := range vs{
		if n>0 {
			result += "\t"
		}
		if i == nil {
			result += "\"\""
		}
		t := reflect.TypeOf(i)
		v := reflect.ValueOf(i)
		result += toString(t,v, false)
	}
	return result
}

func toString(t reflect.Type, v reflect.Value, withQuot bool) string {
	if t == nil {
		return  "nil"
	}
	if t.Kind() == reflect.Ptr {
		if v.IsNil(){
			return "nil"
		}
		v = v.Elem()
		t = v.Type()
		return  toString(t, v, withQuot)
	}
	format := "%s"
	if withQuot {
		format = "\"%s\""
	}
	switch t.Kind() {
	case reflect.Invalid:
		return format
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", v.Float())
	case reflect.String:
		return fmt.Sprintf(format, v.String())
	case reflect.Interface:
		if v.CanInterface() {
			return ToSting(v.Interface())
		}else{
			return "<private>"
		}
	case reflect.Bool:
		return fmt.Sprintf("%t", v.Bool())
	case reflect.Struct:
		return structToString(t, v)
	case reflect.Map:
		return mapToString(t, v)
	case reflect.Array, reflect.Slice:
		return sliceToString(t, v)
	default:
		return "unkown"

	}
	return fmt.Sprintf("%v", v.Interface())
}

func structToString(t reflect.Type, v reflect.Value) string {
	var result bytes.Buffer
	result.WriteString("{")
	for k := 0; k<v.NumField(); k++ {
		if k > 0 {
			result.WriteString(",")
		}
		result.WriteString("\"")
		result.WriteString(t.Field(k).Name)
		result.WriteString("\":")
		result.WriteString(toString(v.Field(k).Type(), v.Field(k), true))
	}
	result.WriteString("}")
	return result.String()
}

func mapToString(t reflect.Type, v reflect.Value) string {
	var result bytes.Buffer
	result.WriteString("{")
	for i,k := range v.MapKeys() {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(toString(k.Type(), k, true))
		result.WriteString(":")
		v1 := v.MapIndex(k)
		result.WriteString(toString(v1.Type(),v1, true))
	}
	result.WriteString("}")
	return result.String()
}

func sliceToString(t reflect.Type, v reflect.Value) string {
	var result bytes.Buffer
	result.WriteString("[")
	for i:=0 ;i<v.Len(); i++ {
		if i > 0 {
			result.WriteString(",")
		}
		result.WriteString(toString(v.Index(i).Type(), v.Index(i), true))
	}
	result.WriteString("]")
	return result.String()
}
