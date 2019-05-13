package mysql

import (
	"fmt"
	"errors"
	"strings"
	"bigbigTravel/component/helper"
)

const (
	vTypeInt			= "int"
	vTypeFloat			= "float"
	vTypeString			= "string"
	vTypeArrayInt		= "arrayInt"
	vTypeArrayString	= "arrayString"
)

type expression struct{
	key			string
	operate     string
	value		interface{}
	vtype		string
}


var operateList = map[string]string{
	"="		: "vector",
	">"		: "vector",
	"<"		: "vector",
	">="	: "vector",
	"<="	: "vector",
	"!="	: "vector",
	"<>"	: "vector",
	"like"	: "string",
	"not like" : "string",
	"in"	: "array",
	"not in"	: "array",
	}

func newExpression(key, operate string, value interface{}, vtype string) *expression {
	if _,ok := operateList[operate]; !ok {
		checkerr(errors.New("sql build failed, invalid operate"))
	}
	exp := &expression{
		key:key,
		operate:operate,
		value:value,
		vtype:vtype,
	}
	return exp
}

func (exp *expression) ToString () string {
	vstr := ""
	switch exp.vtype {
	case vTypeString:
		vstr = fmt.Sprintf("'%s'", RealEscapeString(exp.value.(string)))
	case vTypeFloat:
		vstr = fmt.Sprintf("%f", exp.value)
	case vTypeInt:
		vstr = fmt.Sprintf("%d", exp.value)
	case vTypeArrayInt:
		val := exp.value.([]int)
		if len(val) == 0 {
			checkerr(errors.New("sql build failed, invalid array value"))
		}
		vstr = "(" + helper.JoinIntSlice(val, ",") + ")"
	case vTypeArrayString:
		val := exp.value.([]string)
		if len(val) == 0 {
			checkerr(errors.New("sql build failed, invalid array value"))
		}
		for i:=0; i<len(val) ;i++  {
			val[i] = RealEscapeString(val[i])
		}
		vstr = "('" + strings.Join(val, "','") + "')"
	}
	return fmt.Sprintf("`%s` %s %s", exp.key, exp.operate, vstr)
}
