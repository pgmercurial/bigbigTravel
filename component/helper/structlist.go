package helper

import (
	"reflect"
	"sort"
	"strings"
)

const SortAsc = "asc"
const SortDesc = "desc"

type IStruct interface {
}

type RecordList struct {
	records    []IStruct
	fieldIndex map[string]int
	fieldType  map[string]reflect.Type
	sortKey    string
	isSort     bool
}

func (rs *RecordList) Len() int {
	return len(rs.records)
}

func (rs *RecordList) Less(i, j int) bool {
	ftype, ok := rs.fieldType[rs.sortKey]
	if !ok {
		return i < j
	}
	index := rs.fieldIndex[rs.sortKey]
	v := reflect.ValueOf(rs.records[i])
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	v1 := v.Field(index)
	v = reflect.ValueOf(rs.records[j])
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	v2 := v.Field(index)
	switch {
	case IsInt(ftype):
		return v1.Int() < v2.Int()
	case IsFloat(ftype):
		return v1.Float() < v2.Float()
	case IsString(ftype):
		return v1.String() < v2.String()
	default:
		return 1 < j
	}
}

func (rs *RecordList) Swap(i, j int) {
	rs.isSort = true
	rs.records[i], rs.records[j] = rs.records[j], rs.records[i]
}

func (rs *RecordList) Register(record IStruct) {
	if rs.fieldIndex == nil {
		rs.fieldIndex = make(map[string]int)
		rs.fieldType = make(map[string]reflect.Type)
		t := reflect.TypeOf(record)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		for i := 0; i < t.NumField(); i++ {
			key := strings.ToLower(t.Field(i).Name)
			rs.fieldIndex[key] = i
			rs.fieldType[key] = t.Field(i).Type
		}
	}
}

func (rs *RecordList) Append(record ...IStruct) *RecordList {
	if len(record) == 0 {
		return rs
	}
	rs.records = append(rs.records, record...)
	rs.Register(record[0])
	return rs
}

func (rs *RecordList) AllRecord() []IStruct {
	if rs.Len() == 0 {
		return make([]IStruct, 0)
	}
	return rs.records
}

func (rs *RecordList) FirstRecord() IStruct {
	if len(rs.records) > 0 {
		return rs.records[0]
	} else {
		return nil
	}
}

func (rs *RecordList) LastRecord() IStruct {
	if len(rs.records) > 0 {
		return rs.records[len(rs.records)-1]
	} else {
		return nil
	}
}

func (rs *RecordList) GetAt(index int) IStruct {
	if len(rs.records) > index && index >= 0 {
		return rs.records[index]
	} else {
		return nil
	}

}

func (rs *RecordList) Columns(key string) interface{} {
	key = strings.ToLower(key)
	index, ok := rs.fieldIndex[key]
	if !ok {
		return nil
	}
	if IsInt(rs.fieldType[key]) {
		result := make([]int, 0, rs.Len())
		for _, r := range rs.records {
			v := reflect.ValueOf(r)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			result = append(result, int(v.Field(index).Int()))
		}
		return result
	} else if IsFloat(rs.fieldType[key]) {
		result := make([]float64, 0, rs.Len())
		for _, r := range rs.records {
			v := reflect.ValueOf(r)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			result = append(result, v.Field(index).Float())
		}
		return result
	} else if IsString(rs.fieldType[key]) {
		result := make([]string, 0, rs.Len())
		for _, r := range rs.records {
			v := reflect.ValueOf(r)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			result = append(result, v.Field(index).String())
		}
		return result
	}
	return nil
}

func (rs *RecordList) IndexBy(key string) map[interface{}]IStruct {
	result := make(map[interface{}]IStruct)
	key = strings.ToLower(key)
	index, ok := rs.fieldIndex[key]
	if !ok {
		return result
	}
	for _, r := range rs.records {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		result[v.Field(index).Interface()] = r
	}
	return result
}

// 将传入的key作为返回map的key，如果key出现多次则进行追加返回
func (rs *RecordList) IndexSliceBy(key string) map[interface{}][]IStruct {
	result := make(map[interface{}][]IStruct)
	key = strings.ToLower(key)
	index, ok := rs.fieldIndex[key]
	if !ok {
		return result
	}
	for _, r := range rs.records {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		key := v.Field(index).Interface()
		value := make([]IStruct, 0)
		if slice, ok := result[key]; ok {
			value = append(slice, r)
		} else {
			value = append(value, r)
		}
		result[key] = value
	}
	return result
}

func (rs *RecordList) GroupBy(key string) map[interface{}][]IStruct {
	result := make(map[interface{}][]IStruct)
	key = strings.ToLower(key)
	index, ok := rs.fieldIndex[key]
	if !ok {
		return result
	}
	for _, r := range rs.records {
		v := reflect.ValueOf(r)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		mkey := v.Field(index).Interface()
		if _, ok := result[mkey]; !ok {
			result[mkey] = make([]IStruct, 0)
		}
		result[mkey] = append(result[mkey], r)
	}
	return result
}

func (rs *RecordList) SortBy(key string, order string) *RecordList {
	rs.isSort = false
	switch order {
	case SortAsc:
		rs.sortKey = strings.ToLower(key)
		sort.Sort(rs)
		return rs
	case SortDesc:
		rs.sortKey = strings.ToLower(key)
		trs := sort.Reverse(rs)
		sort.Sort(trs)
		return rs
	default:
		return rs
	}
}

func (rs *RecordList) IsSort() bool {
	return rs.isSort
}
