package mysql


import (
	"database/sql"
	"reflect"
	"strings"
	"bigbigTravel/component/helper"
)

type QueryStat struct {
	affectRows	int
	lastInsertId	int
	rows *sql.Rows
	record ActiveRecord
}

func (stat *QueryStat) Fetch(record ...ActiveRecord) ActiveRecord {
	defer stat.Close()
	if len(record) > 0 {
		return stat.fetchOne(record[0])
	}else {
		return stat.fetchOne(nil)
	}
}

func (stat *QueryStat) fetchOne(record ActiveRecord) ActiveRecord {
	if stat.rows == nil {
		return nil
	}
	if stat.rows.Next() {
		t := reflect.TypeOf(stat.record)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		var result reflect.Value
		if record != nil {
			result = reflect.ValueOf(record)
		}else {
			result = reflect.New(t)
		}
		v := reflect.Indirect(result)
		var m = map[string]interface{}{}
		for k := 0; k < v.NumField(); k++ {
			dbField := t.Field(k).Tag.Get("column")
			if dbField == "" {
				continue
			}
			m[strings.ToLower(dbField)] = v.Field(k).Addr().Interface()
		}
		//防止数据结构中定义的元素先后顺序与MySQL中不一致，要取rows中的顺序映射一遍
		types,err := stat.rows.ColumnTypes()
		checkerr(err)
		args := make([]interface{}, 0, len(types))
		for _,t := range types{
			if pointer,ok := m[strings.ToLower(t.Name())]; ok{
				args = append(args, pointer)
			}
		}
		//将固定的某数据结构中的各个元素的地址按声明顺序依次放入slice中，再用scan方法为其赋值
		err = stat.rows.Scan(args...)
		checkerr(err)
		return result.Interface().(ActiveRecord)
	}else{
		return nil
	}
}

func (stat *QueryStat) FetchAll() *helper.RecordList {
	result := new(helper.RecordList)
	result.Register(stat.record)
	for {
		record := stat.fetchOne(nil)
		if record == nil {
			return result
		}
		result.Append(record)
	}
	return result
}

func (stat *QueryStat) Rows () *sql.Rows {
	return stat.rows
}

func (stat *QueryStat) LastInsertId () int {
	return stat.lastInsertId
}

func (stat *QueryStat) AffectRows () int {
	return stat.affectRows
}

func (stat *QueryStat) Close () {
	stat.rows.Close()
}