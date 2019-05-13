package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	"bigbigTravel/component/logger"
)

type DeleteBuilder struct {
	identify string
	db       interface{}
	context  *Context
	rowSql   string
	qSql     *deleteSql
	record   ActiveRecord
}

type deleteSql struct {
	table  string
	params []interface{}
	where  *Condition
	limit  int
}

func NewDeleteBuilder(identify string, db interface{}, context *Context, r ActiveRecord) *DeleteBuilder {
	builder := &DeleteBuilder{
		identify: identify,
		db:       db,
		context:  context,
		record:   r,
		qSql: &deleteSql{
			table:  r.Name(),
			params: make([]interface{}, 0),
			where: &Condition{
				logic: logicAnd,
			},
		},
	}
	builder.qSql.where.fields = getRecordFields(r.Name(), identify)
	return builder
}

func (builder *DeleteBuilder) Where(key string, operate string, value interface{}) *DeleteBuilder {
	builder.qSql.where.AddExpress(logicAnd, key, operate, value)
	return builder
}

func (builder *DeleteBuilder) AndWhere(key string, operate string, value interface{}) *DeleteBuilder {
	builder.qSql.where.AddExpress(logicAnd, key, operate, value)
	return builder
}

func (builder *DeleteBuilder) OrWhere(key string, operate string, value interface{}) *DeleteBuilder {
	builder.qSql.where.AddExpress(logicOr, key, operate, value)
	return builder
}

func (builder *DeleteBuilder) Limit(limit int) *DeleteBuilder {
	builder.qSql.limit = limit
	return builder
}

func (builder *DeleteBuilder) CreateCommand() *DeleteBuilder {
	if builder.qSql.table == "" {
		checkerr(errors.New("bad sql, no table"))
	}
	where := builder.qSql.where.ToString()
	if where == "" {
		checkerr(errors.New("bad sql, delete should set conditions"))
	}
	builder.rowSql = "DELETE FROM `" + builder.qSql.table + "` WHERE " + where
	if builder.qSql.limit != 0 {
		builder.rowSql += fmt.Sprintf(" LIMIT %d", builder.qSql.limit)
	}
	return builder
}

func (builder *DeleteBuilder) Execute() int {
	if builder.rowSql == "" {
		builder.CreateCommand()
	}
	var stmt *sql.Stmt
	var err error
	if db, ok := builder.db.(*sql.DB); ok && db != nil {
		logger.Debug("sql", builder.context.RequestId, "normal", builder.rowSql)
		stmt, err = db.Prepare(builder.rowSql)
	} else if db, ok := builder.db.(*sql.Tx); ok && db != nil {
		logger.Debug("sql", builder.context.RequestId, "transaction:"+builder.context.TxId, builder.rowSql)
		stmt, err = db.Prepare(builder.rowSql)
	} else {
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
