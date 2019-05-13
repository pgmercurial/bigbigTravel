package mysql

import (
	"database/sql"
	"errors"
	"strings"

	"fmt"

	"bigbigTravel/component/logger"
)

type InsertBuilder struct {
	db      interface{}
	context *Context
	rowSql  string
	qSql    *insertSql
	record  ActiveRecord
}

type insertSql struct {
	table       string
	columns     []string
	records     []ActiveRecord
	value       []string
	values      []string
	notExistSql string
	params      []interface{}
}

func NewInsertBuilder(db interface{}, context *Context, r ActiveRecord) *InsertBuilder {
	builder := &InsertBuilder{
		db:      db,
		context: context,
		record:  r,
		qSql: &insertSql{
			table:   r.Name(),
			columns: make([]string, 0, 5),
			records: make([]ActiveRecord, 0),
			values:  make([]string, 0),
			params:  make([]interface{}, 0),
		},
	}
	return builder
}

func (builder *InsertBuilder) Columns(keys ...string) *InsertBuilder {
	if len(keys) == 0 {
		checkerr(errors.New("sql build failed, insert need at least one column"))
	}
	builder.qSql.columns = keys
	for i := 0; i < len(builder.qSql.columns); i++ {
		builder.qSql.value = append(builder.qSql.value, "?")
	}
	return builder
}

func (builder *InsertBuilder) AddRecords(r ...ActiveRecord) *InsertBuilder {
	builder.qSql.records = append(builder.qSql.records, r...)
	return builder
}

func (builder *InsertBuilder) NotExist(notExistSql string) *InsertBuilder {
	builder.qSql.notExistSql = notExistSql
	return builder
}

func (builder *InsertBuilder) Value(value ...interface{}) *InsertBuilder {
	if len(value) == 0 || len(value) != len(builder.qSql.columns) {
		checkerr(errors.New("sql build failed, add values nums not equal the columns"))
	}

	builder.qSql.values = append(builder.qSql.values, "("+strings.Join(builder.qSql.value, ",")+")")
	builder.qSql.params = append(builder.qSql.params, value...)
	return builder
}

func (builder *InsertBuilder) Values(values ...[]interface{}) *InsertBuilder {
	for _, value := range values {
		builder.Value(value...)
	}
	return builder
}

func (builder *InsertBuilder) CreateCommand() *InsertBuilder {
	if builder.qSql.table == "" {
		checkerr(errors.New("bad sql, no table"))
	}
	if len(builder.qSql.columns) == 0 {
		checkerr(errors.New("bad sql, no columns"))
	}
	builder.rowSql = "INSERT IGNORE INTO `" + builder.qSql.table + "` (`" + strings.Join(builder.qSql.columns, "`,`") + "`)"
	//+ " VALUES " + strings.Join(builder.qSql.values, ",")

	if builder.qSql.notExistSql != "" {
		var fields string
		length := len(builder.qSql.params)
		switch length {
		case 1:
			fields = "?"
		case 2:
			fields = "?, ?"
		case 0:
		default:
			fields = "?,"
			for i := 1; i < length-1; i++ {
				fields = fmt.Sprintf("%s ?, ", fields)
			}
			fields = fmt.Sprintf("%s?", fields)
		}
		builder.rowSql = fmt.Sprintf("%s SELECT %s FROM dual WHERE not exists (%s)", builder.rowSql, fields, builder.qSql.notExistSql)
	} else {
		builder.rowSql = fmt.Sprintf("%s %s %s", builder.rowSql, "VALUES", strings.Join(builder.qSql.values, ","))
	}

	return builder
}

func (builder *InsertBuilder) Execute() *QueryStat {
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
	var res sql.Result
	//if builder.qSql.notExistSql != "" {
	//	res, err = stmt.Exec()
	//} else {
	res, err = stmt.Exec(builder.qSql.params...)
	//}
	checkerr(err)
	rows, err := res.RowsAffected()
	checkerr(err)
	lastInsertId, err := res.LastInsertId()
	checkerr(err)
	return &QueryStat{
		affectRows:   int(rows),
		lastInsertId: int(lastInsertId),
	}
}
