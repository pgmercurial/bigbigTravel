package mysql

import (
	"database/sql"
	"fmt"
	"errors"
	"reflect"
	"bigbigTravel/component/helper"
	"strings"
	"bigbigTravel/component/logger"
)

type UpdateBuilder struct {
	identify 	string
	db			interface{}
	tx 			*sql.Tx
	context		*Context
	rowSql		string
	qSql		*updateSql
	record 		ActiveRecord
}

type updateSql struct {
	table		string
	sets		[]string
	params      []interface{}
	where		*Condition
	limit		int
}

func NewUpdateBuilder(identify 	string, db interface{},  context *Context, r ActiveRecord) *UpdateBuilder {
	builder := &UpdateBuilder{
		identify:identify,
		db:db,
		context:context,
		record:r,
		qSql:&updateSql{
			table:r.Name(),
			sets: make([]string, 0),
			params:make([]interface{}, 0),
			where:&Condition{
				logic:logicAnd,
			},
		},
	}
	builder.qSql.where.fields = getRecordFields(r.Name(), identify)
	return builder
}

func (builder *UpdateBuilder) Set (key string, value interface{}) *UpdateBuilder {
	t := reflect.TypeOf(value)
	switch {
	case helper.IsInt(t):
		builder.qSql.sets = append(builder.qSql.sets, fmt.Sprintf("`%s` = %d", key, value))
	case helper.IsFloat(t):
		builder.qSql.sets = append(builder.qSql.sets, fmt.Sprintf("`%s` = %f", key, value))
	case helper.IsString(t):
		builder.qSql.sets = append(builder.qSql.sets, fmt.Sprintf("`%s` = ?", key))
		builder.qSql.params = append(builder.qSql.params, value)
	default:
		checkerr(errors.New("invalid set value type, should be numeric or string"))
	}
	return builder
}

func (builder *UpdateBuilder) Inc (key string, value interface{}) *UpdateBuilder {
	k := reflect.TypeOf(value)
	switch {
	case helper.IsInt(k):
		o := "+"
		if value.(int) <= 0 {
			o = "-"
			value = -value.(int)
		}
		builder.qSql.sets = append(builder.qSql.sets, fmt.Sprintf("`%s` = `%s` %s %d", key, key, o, value))
	case helper.IsFloat(k):
		o := "+"
		if value.(float64) <= 0 {
			o = "-"
			value = -value.(float64)
		}
		builder.qSql.sets = append(builder.qSql.sets, fmt.Sprintf("`%s` = `%s` %s %f", key, key, o, value))
	default:
		checkerr(errors.New("invalid set value type, should be numeric or string"))
	}
	return builder
}

func (builder *UpdateBuilder) Where (key string, operate string, value interface{}) *UpdateBuilder {
	builder.qSql.where.AddExpress(logicAnd,key,operate,value)
	return builder
}

func (builder *UpdateBuilder) AndWhere (key string, operate string, value interface{}) *UpdateBuilder {
	builder.qSql.where.AddExpress(logicAnd,key,operate,value)
	return builder
}

func (builder *UpdateBuilder) OrWhere (key string, operate string, value interface{}) *UpdateBuilder {
	builder.qSql.where.AddExpress(logicOr,key,operate,value)
	return builder
}

func (builder *UpdateBuilder) Limit (limit int) *UpdateBuilder {
	builder.qSql.limit = limit
	return builder
}

func (builder *UpdateBuilder) CreateCommand() *UpdateBuilder {
	if len(builder.qSql.sets) == 0{
		checkerr(errors.New("bad sql, empty sets"))
	}
	if builder.qSql.table == ""{
		checkerr(errors.New("bad sql, no table"))
	}
	where := builder.qSql.where.ToString()
	if where == ""{
		checkerr(errors.New("bad sql, update should set conditions"))
	}
	set := strings.Join(builder.qSql.sets, ",")
	builder.rowSql = "UPDATE `"+ builder.qSql.table +"` SET "+ set+ " WHERE "+ where
	if builder.qSql.limit != 0 {
		builder.rowSql += fmt.Sprintf(" LIMIT %d", builder.qSql.limit)
	}
	return builder
}

func (builder *UpdateBuilder) Execute() int {
	if builder.rowSql == "" {
		builder.CreateCommand()
	}
	var stmt *sql.Stmt
	var err error
	if db,ok := builder.db.(*sql.DB); ok && db != nil{
		logger.Debug("sql", builder.context.RequestId, "normal", builder.rowSql)
		stmt,err = db.Prepare(builder.rowSql)
	}else if db,ok := builder.db.(*sql.Tx); ok && db != nil{
		logger.Debug("sql", builder.context.RequestId, "transaction:"+builder.context.TxId,  builder.rowSql)
		stmt,err = db.Prepare(builder.rowSql)
	}else {
		checkerr(errors.New("expect db or tx to execute sql"))
	}

	defer stmt.Close()

	checkerr(err)
	res, err := stmt.Exec(builder.qSql.params...)
	checkerr(err)
	num, err := res.RowsAffected()
	checkerr(err)
	return int(num)
}
