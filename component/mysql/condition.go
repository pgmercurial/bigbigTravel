package mysql

import (
	"strings"
	"errors"
	"reflect"
	"fmt"
	"strconv"
)

const logicAnd  = "AND"
const logicOr   = "OR"

type Condition struct {
	logic		string
	fields		map[string]*RecordField
	expressions	[]*expression
	conditions	[]*Condition
}

func (condition *Condition) AddExpress(logicType, key, operate string, value interface{})  {
	if ot, ok := operateList[operate]; !ok {
		checkerr(errors.New("sql build failed, invalid operate"))
	}else {
		vt := reflect.TypeOf(value)
		fieldInfo, ok := condition.fields[key]
		switch ot {
		case "vector" :
			switch vt.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if ok && fieldInfo.Type == FieldTypeString {
					condition.addExpress(logicType,key,operate,fmt.Sprintf("%d", value), vTypeString)
				}else {
					condition.addExpress(logicType,key,operate,value, vTypeInt)
				}
			case reflect.String:
				if ok && fieldInfo.Type == FieldTypeInt {
					v, err := strconv.ParseInt(value.(string), 10, 64)
					checkerr(err)
					condition.addExpress(logicType,key,operate, v, vTypeInt)
				}else {
					condition.addExpress(logicType, key, operate, value, vTypeString)
				}
			case reflect.Float32, reflect.Float64:
				condition.addExpress(logicType,key,operate,value, vTypeFloat)
			default:
				checkerr(errors.New("sql build failed, invalid operate"))
			}
		case "string" :
			switch vt.Kind() {
			case reflect.String:
				condition.addExpress(logicType,key,operate,value, vTypeString)
			default:
				checkerr(errors.New("sql build failed, invalid operate"))
			}
		case "array":
			v := reflect.TypeOf(value).String()
			switch v {
			case "[]int":
				if ok && fieldInfo.Type == FieldTypeString {
					newV := make([]string, 0, len(value.([]int)))
					for _,s := range value.([]int){
						newV = append(newV, fmt.Sprintf("%d", s))
					}
					condition.addExpress(logicType,key,operate,newV, vTypeArrayString)
				}else{
					condition.addExpress(logicType, key, operate, value, vTypeArrayInt)
				}
			case "[]string":
				if ok && fieldInfo.Type == FieldTypeInt{
					newV := make([]int, 0, len(value.([]string)))
					for _,s := range value.([]string) {
						vint, err := strconv.ParseInt(s, 10, 64)
						if err != nil{
							checkerr(errors.New("sql build failed, "+err.Error()))
						}

						newV = append(newV, int(vint))
					}
					condition.addExpress(logicType,key,operate,newV, vTypeArrayInt)
				}else {
					condition.addExpress(logicType,key,operate,value, vTypeArrayString)
				}
			default:
				checkerr(errors.New("sql build failed, invalid operate"))
			}
		}
	}
}

func (condition *Condition) addExpress(logicType, key, operate string, value interface{}, vType string) *Condition {
	if logicType != condition.logic && condition.logic != ""{
		oldCondition := new(Condition)
		*oldCondition = *condition
		condition.conditions = []*Condition{oldCondition}
		condition.logic = logicType
		condition.expressions = []*expression{
				newExpression(key,operate,value,vType),
		}
		return condition
	}else {
		condition.expressions = append(condition.expressions, newExpression(key,operate,value,vType))
		return condition
	}
}

func (condition *Condition) ToString() string {
	conStr := ""
	for _,c := range condition.conditions{
		conStr += "("+c.ToString()+")"
	}
	expStrArr := make([]string,0,len(condition.expressions))
	for _,e := range condition.expressions {
		expStrArr = append(expStrArr, e.ToString())
	}
	if conStr != "" {
		conStr += " "+condition.logic+" "
	}
	if len(expStrArr) > 0 {
		conStr += "(" + strings.Join(expStrArr, " "+condition.logic+" ") +")"
	}
	return conStr
}

func generateValue()  {

}

