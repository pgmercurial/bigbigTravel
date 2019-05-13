package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"bigbigTravel/component/logger"
)

type QueryBuilder struct {
	identify string
	db       interface{}
	context  *Context
	rowSql   string
	qSql     *querySql
	record   ActiveRecord
}

type querySql struct {
	table   string
	columns []string
	funcCol []string
	where   *Condition
	orderBy string
	offset  int
	limit   int
}

func NewQueryBuilder(identify string, db interface{}, context *Context, r ActiveRecord) *QueryBuilder {
	builder := &QueryBuilder{
		identify: identify,
		db:       db,
		context:  context,
		record:   r,
		qSql: &querySql{
			table: r.Name(),
			where: &Condition{
				logic: logicAnd,
			},
		},
	}
	builder.qSql.where.fields = getRecordFields(r.Name(), identify)
	return builder
}

func NewQueryFinder(identify string, db interface{}, context *Context, f Finder) *QueryBuilder {
	builder := &QueryBuilder{
		identify: identify,
		db:       db,
		context:  context,
		record:   f.Record(),
		qSql: &querySql{
			table: f.Record().Name(),
			where: &Condition{
				logic: logicAnd,
			},
		},
	}
	builder.qSql.where.fields = f.Fields()
	return builder
}

func (builder *QueryBuilder) Select(column ...string) *QueryBuilder {
	for _, c := range column {
		if c == "*" {
			builder.qSql.columns = make([]string, 0, len(builder.qSql.where.fields))
			for _, field := range builder.qSql.where.fields {
				builder.qSql.columns = append(builder.qSql.columns, field.Column)
			}
			break
		}
	}
	if builder.qSql.columns == nil {
		builder.qSql.columns = column
	}
	return builder
}

func (builder *QueryBuilder) From(table string) *QueryBuilder {
	builder.qSql.table = table
	builder.record = getRegisterRecord(table, builder.identify)
	return builder
}

func (builder *QueryBuilder) Where(key string, operate string, value interface{}) *QueryBuilder {
	builder.qSql.where.AddExpress(logicAnd, key, operate, value)
	return builder
}

func (builder *QueryBuilder) AndWhere(key string, operate string, value interface{}) *QueryBuilder {
	builder.qSql.where.AddExpress(logicAnd, key, operate, value)
	return builder
}

func (builder *QueryBuilder) OrWhere(key string, operate string, value interface{}) *QueryBuilder {
	builder.qSql.where.AddExpress(logicOr, key, operate, value)
	return builder
}

func (builder *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {
	builder.qSql.orderBy = orderBy
	return builder
}

func (builder *QueryBuilder) Offset(offset int) *QueryBuilder {
	builder.qSql.offset = offset
	return builder
}

func (builder *QueryBuilder) Limit(limit int) *QueryBuilder {
	builder.qSql.limit = limit
	return builder
}

func (builder *QueryBuilder) CreateCommand() *QueryBuilder {
	if len(builder.qSql.columns) == 0 {
		checkerr(errors.New("bad sql, need identify select fields"))
	}
	if builder.qSql.table == "" {
		checkerr(errors.New("bad sql, miss table name"))
	}
	where := builder.qSql.where.ToString()
	selectStr := make([]string, 0, len(builder.qSql.columns))
	for _, sl := range builder.qSql.columns {
		if strings.Contains(sl, " ") || strings.Contains(sl, "(") {
			selectStr = append(selectStr, sl)
		} else {
			selectStr = append(selectStr, "`"+sl+"`")
		}
	}

	builder.rowSql = "SELECT " + strings.Join(selectStr, ",") + " FROM `" + builder.qSql.table + "`"
	if where != "" {
		builder.rowSql += " WHERE " + builder.qSql.where.ToString()
	}
	if strings.Trim(builder.qSql.orderBy, " ") != "" {
		builder.rowSql += " ORDER BY " + builder.qSql.orderBy
	}
	if builder.qSql.limit != 0 {
		builder.rowSql += fmt.Sprintf(" LIMIT %d,%d", builder.qSql.offset, builder.qSql.limit)
	}
	return builder
}

func (builder *QueryBuilder) Count(column ...string) int {
	if len(column) == 0 {
		builder.qSql.columns = []string{"count(1) as count"}
	} else {
		builder.qSql.columns = column
	}

	if builder.rowSql == "" {
		builder.CreateCommand()
	}
	var rows *sql.Rows
	var err error
	if db, ok := builder.db.(*sql.DB); ok && db != nil {
		logger.Debug("sql", builder.context.RequestId, "normal", builder.rowSql)
		rows, err = db.Query(builder.rowSql)
		defer rows.Close()
	} else if db, ok := builder.db.(*sql.Tx); ok && db != nil {
		logger.Debug("sql", builder.context.RequestId, "transaction:"+builder.context.TxId, builder.rowSql)
		rows, err = db.Query(builder.rowSql)
		defer rows.Close()
	} else {
		checkerr(errors.New("expect db or tx to execute sql"))
	}
	checkerr(err)
	var count int
	if rows.Next() {
		err = rows.Scan(&count)
		checkerr(err)
		return count
	}
	checkerr(errors.New("count query error"))
	return 0
}

func (builder *QueryBuilder) Execute() *QueryStat {
	if builder.rowSql == "" {
		builder.CreateCommand()
	}
	var rows *sql.Rows
	var err error
	if db, ok := builder.db.(*sql.DB); ok && db != nil {
		logger.Debug("sql", builder.context.RequestId, "normal", builder.rowSql)
		rows, err = db.Query(builder.rowSql)
	} else if db, ok := builder.db.(*sql.Tx); ok && db != nil {
		logger.Debug("sql", builder.context.RequestId, "transaction:"+builder.context.TxId, builder.rowSql)
		rows, err = db.Query(builder.rowSql)
	} else {
		checkerr(errors.New("expect db or tx to execute sql"))
	}
	checkerr(err)
	stat := &QueryStat{rows: rows, record: builder.record}
	return stat
}
