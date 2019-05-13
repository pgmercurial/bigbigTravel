package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"bigbigTravel/component/exception"
	"bigbigTravel/component/helper"
	"bigbigTravel/component/logger"
)

type RecordField struct {
	Index     int
	Key       string
	Column    string
	Type      string
	IsPrimary bool
	Modify    bool
}

func RegisterRecord(name string, activeRecord ActiveRecord, identify ...string) {
	if _, ok := registerRecords[name]; ok {
		return
	}
	insName := DefaultInstance
	if len(identify) > 0 {
		insName = identify[0]
	}
	if _, ok := registerRecords[insName]; !ok {
		registerRecords[insName] = make(map[string]map[string]interface{})
	}
	registerRecords[insName][name] = make(map[string]interface{})
	registerRecords[insName][name]["model"] = activeRecord
	t := reflect.TypeOf(activeRecord).Elem()
	registerRecords[insName][name]["fields"] = make(map[string]*RecordField)
	for i := 0; i < t.NumField(); i++ {
		recordField := new(RecordField)
		recordField.Index = i
		recordField.Key = t.Field(i).Name

		recordField.Column = t.Field(i).Tag.Get("column")
		recordField.Type = t.Field(i).Type.Name()
		recordField.Modify = t.Field(i).Tag.Get("modify") != "false"
		_, recordField.IsPrimary = t.Field(i).Tag.Lookup("primary")
		if recordField.Column == "" {
			checkerr(errors.New("record register " + name + " failed, need column tag"))
		}
		if recordField.Type == "" {
			recordField.Type = FieldTypeString
		}
		registerRecords[insName][name]["fields"].(map[string]*RecordField)[recordField.Column] = recordField
		if recordField.IsPrimary {
			if _, ok := registerRecords[name]["primary"]; ok {
				checkerr(errors.New("record register " + name + " failed, too many primary"))
			}
			registerRecords[insName][name]["primary"] = recordField
		}
	}
}

func getRegisterRecord(table string, identify ...string) ActiveRecord {
	insName := DefaultInstance
	if len(identify) > 0 {
		insName = identify[0]
	}
	record, ok := registerRecords[insName][table]["model"]
	if !ok {
		checkerr(errors.New("table not found in register record list"))
	}
	return record.(ActiveRecord)
}

func getRecordFields(name string, identify ...string) map[string]*RecordField {
	insName := DefaultInstance
	if len(identify) > 0 {
		insName = identify[0]
	}
	if fields, ok := registerRecords[insName][name]["fields"]; ok {
		return fields.(map[string]*RecordField)
	}
	return make(map[string]*RecordField)
}

func getRecordField(recordName string, fieldKey string, identify ...string) *RecordField {
	insName := DefaultInstance
	if len(identify) > 0 {
		insName = identify[0]
	}
	fields, ok := registerRecords[insName][recordName]["fields"]
	if !ok {
		return nil
	}
	fs := fields.(map[string]*RecordField)
	if f, ok := fs[fieldKey]; ok {
		return f
	}
	for _, v := range fs {
		if strings.ToLower(v.Column) == strings.ToLower(fieldKey) || strings.ToLower(fieldKey) == strings.ToLower(v.Key) {
			return v
		}
	}
	return nil
}

func getRecordPrimary(table string, identify ...string) *RecordField {
	insName := DefaultInstance
	if len(identify) > 0 {
		insName = identify[0]
	}
	field, ok := registerRecords[insName][table]["primary"]
	if !ok {
		checkerr(errors.New("table did not set primary key, check the struct defined"))
	}
	return field.(*RecordField)
}

//
type Context struct {
	RequestId string
	TxId      string //事物id
	UseCache  bool
}

type Finder interface {
	Record() ActiveRecord
	Fields() map[string]*RecordField
}

type IDB interface {
	Query(recordName, sql string, params ...interface{}) *helper.RecordList
	FindOneByPrimary(table string, id int) ActiveRecord
	FindListByPrimary(table string, ids []int) *helper.RecordList
	Find(table string) *QueryBuilder
	Finder(f Finder) *QueryBuilder
	Update(table string) *UpdateBuilder
	Delete(table string) *DeleteBuilder
	Insert(table string) *InsertBuilder
	BatchInsert(rs []ActiveRecord, columns ...string) int
	SaveRecord(r ActiveRecord, columns ...string) int
	LoadRecord(r ActiveRecord) bool
	DeleteRecord(r ActiveRecord) bool
	//SearchRecord(r ActiveRecord, columns ...string) *helper.RecordList
}

////////////////////////db连接////////////////////////////
type Cache interface {
	Set(key string, value interface{})
	Get(key string) string
	SetMulti(kv map[string]interface{})
	GetMulti(key ...string) map[string]string
}

type CacheFactory func() Cache

type MysqlDB struct {
	Identify string
	db       *sql.DB
	slavers  []*sql.DB
	context  *Context
	cache    Cache
	useCache bool
}

type MysqlConfig struct {
	Host            string
	Port            int64
	UserName        string
	PassWord        string
	DBName          string
	Slaves          []string
	MaxConnLifeTime int64
	MaxIdleConns    int64
	MaxOpenConns    int64
}

var _ IDB = &MysqlDB{}

var instance = make(map[string]*sql.DB)
var registerRecords = map[string]map[string]map[string]interface{}{}
var config = make(map[string]*MysqlConfig)
var defaultContext = &Context{RequestId: "-"}
var cacheFactory CacheFactory = func() Cache {
	return nil
}
var DefaultInstance = "default"

func InitConfig(cfg *MysqlConfig, identify ...string) {
	insName := DefaultInstance
	if len(identify) > 0 {
		insName = identify[0]
	}
	config[insName] = cfg
	instance[insName] = open(cfg)
	for i, slv := range cfg.Slaves {
		slvcfg := new(MysqlConfig)
		arr := strings.Split(slv, ":")
		slvcfg.Host = strings.Trim(arr[0], " ")
		slvcfg.Port = 3306
		if len(arr) == 2 {
			p, err := strconv.Atoi(arr[1])
			if err == nil {
				slvcfg.Port = int64(p)
			}
		}
		slvcfg.DBName = cfg.DBName
		slvcfg.MaxConnLifeTime = cfg.MaxConnLifeTime
		slvcfg.MaxIdleConns = cfg.MaxIdleConns
		slvcfg.MaxOpenConns = cfg.MaxOpenConns
		slvcfg.PassWord = cfg.PassWord
		slvcfg.UserName = cfg.UserName
		slaveName := getSlaveInsName(insName, i)
		instance[slaveName] = open(slvcfg)
		config[slaveName] = slvcfg
	}

}

func getSlaveInsName(identify string, index int) string {
	return fmt.Sprintf("slave:%s%d", identify, index)
}

func SetCache(c CacheFactory) {
	cacheFactory = c
}

func GetInstance(useCache bool, context ...*Context) (db *MysqlDB) {
	if ins, ok := instance[DefaultInstance]; !ok || ins == nil {
		instance[DefaultInstance] = open(config[DefaultInstance])

	}
	db = &MysqlDB{
		Identify: DefaultInstance,
		db:       instance[DefaultInstance],
		context:  defaultContext,
		cache:    cacheFactory(),
		useCache: useCache,
	}
	if len(context) > 0 && context[0] != nil {
		db.context = context[0]
	}
	for i := 0; ; i++ {
		slaveName := getSlaveInsName(DefaultInstance, i)
		slaveDb, ok := instance[slaveName]
		if !ok || slaveDb == nil {
			if cfg, o := config[slaveName]; o {
				slaveDb = open(cfg)
				instance[slaveName] = slaveDb
			} else {
				break
			}
		}
		db.slavers = append(db.slavers, slaveDb)
	}
	return db
}

func GetInsByIdentify(identify string, context ...*Context) (db *MysqlDB) {
	if ins, ok := instance[identify]; !ok || ins == nil {
		instance[identify] = open(config[identify])

	}
	db = &MysqlDB{
		Identify: identify,
		db:       instance[identify],
		context:  defaultContext,
		cache:    cacheFactory(),
	}
	if len(context) > 0 && context[0] != nil {
		db.context = context[0]
		db.useCache = context[0].UseCache
	}
	for i := 0; ; i++ {
		slaveName := getSlaveInsName(identify, i)
		slaveDb, ok := instance[slaveName]
		if !ok || slaveDb == nil {
			if cfg, o := config[slaveName]; o {
				slaveDb = open(cfg)
				instance[slaveName] = slaveDb
			} else {
				break
			}
		}
		db.slavers = append(db.slavers, slaveDb)
	}
	return db
}

func open(config *MysqlConfig) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.UserName,
		config.PassWord,
		config.Host,
		config.Port,
		config.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	checkerr(err)
	if config.MaxConnLifeTime > 0 {
		db.SetConnMaxLifetime(time.Duration(config.MaxConnLifeTime))
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(int(config.MaxIdleConns))
	}
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(int(config.MaxOpenConns))
	}
	//set log
	go getCurrentMysqlConnStatus(db, config.Host, config.MaxOpenConns, config.MaxIdleConns, config.MaxConnLifeTime)
	return db
}

func getCurrentMysqlConnStatus(db *sql.DB, host string, maxConn, maxIdle, lifeTime int64) {
	timer := time.NewTimer(60 * time.Second) // 新建定时器
	for {
		select {
		case <-timer.C:
			timer.Reset(60 * time.Second) // 重新设置定时器
			var stat sql.DBStats
			stat = db.Stats()
			logger.Info("mysql-conn", helper.GenerateUUID(), host, map[string]int64{
				"当前建立连接数":   int64(stat.OpenConnections), //目前的连接总数: 包括空闲及使用中
				"连接存活周期(s)": lifeTime,
				"最大建立连接数":   maxConn,
				"最大空闲连接数":   maxIdle,
			})
		}
	}
}

func (mysql *MysqlDB) masterDB() *sql.DB {
	//fmt.Println("use master")
	return mysql.db
}

func (mysql *MysqlDB) slaveDB() *sql.DB {
	if len(mysql.slavers) == 0 {
		return mysql.db
	}
	index := rand.Intn(len(mysql.slavers))
	//fmt.Println("use slave", index)
	return mysql.slavers[index]
}

func (mysql *MysqlDB) Query(recordName, sql string, params ...interface{}) *helper.RecordList {
	record := getRegisterRecord(recordName, mysql.Identify)
	rows, err := mysql.slaveDB().Query(sql, params...)
	logger.Debug("sql", mysql.context.RequestId, sql)
	checkerr(err)
	stat := &QueryStat{rows: rows, record: record}
	return stat.FetchAll()
}

// func (mysql *MysqlDB) AllCount(recordName string) int {
// 	record := getRegisterRecord(recordName)
// 	queryBuilder := NewQueryBuilder(mysql.db,  mysql.context, record.(ActiveRecord))

// 	return queryBuilder.Count()
// }

func (mysql *MysqlDB) AllCount(recordName string, conditions ...interface{}) int {
	record := getRegisterRecord(recordName, mysql.Identify)
	queryBuilder := NewQueryBuilder(mysql.Identify, mysql.slaveDB(), mysql.context, record.(ActiveRecord))
	if len(conditions) > 0 {
		mp := conditions[0].(map[string]string)
		expressions := make([]*expression, 0, len(mp))
		for k, v := range mp {
			f := getRecordField(recordName, k, mysql.Identify)
			if f == nil {
				continue
			}
			eps, err := getExpresstions(f.Column, f.Type, v)
			if err != nil {
				exception.Panic(exception.ExceptionInvalidParams, err)
			}
			expressions = append(expressions, eps...)
		}
		for _, ep := range expressions {
			queryBuilder.Where(ep.key, ep.operate, ep.value)
		}
	}
	return queryBuilder.Count()
}

func (mysql *MysqlDB) FindOneByPrimary(table string, id int) ActiveRecord {
	record := getRegisterRecord(table, mysql.Identify)
	primaryField := getRecordPrimary(table, mysql.Identify)
	queryBuilder := NewQueryBuilder(mysql.Identify, mysql.slaveDB(), mysql.context, record.(ActiveRecord))
	stat := queryBuilder.Select("*").Where(primaryField.Column, "=", id).Execute().Fetch()
	return stat
}

func (mysql *MysqlDB) FindListByPrimary(table string, ids []int) *helper.RecordList {
	realIds := make([]int, 0, len(ids))
	for _, id := range ids {
		if id != 0 {
			realIds = append(realIds, id)
		}
	}
	if len(realIds) == 0 {
		return new(helper.RecordList)
	}
	r := getRegisterRecord(table, mysql.Identify)
	primaryField := getRecordPrimary(table, mysql.Identify)
	queryBuilder := NewQueryBuilder(mysql.Identify, mysql.slaveDB(), mysql.context, r.(ActiveRecord))
	recordList := queryBuilder.Select("*").Where(primaryField.Column, "in", realIds).Execute().FetchAll()
	return recordList
}

func (mysql *MysqlDB) Find(table string) *QueryBuilder {
	record := getRegisterRecord(table, mysql.Identify)
	queryBuilder := NewQueryBuilder(mysql.Identify, mysql.slaveDB(), mysql.context, record.(ActiveRecord))
	return queryBuilder
}

func (mysql *MysqlDB) Finder(f Finder) *QueryBuilder {
	queryBuilder := NewQueryFinder(mysql.Identify, mysql.slaveDB(), mysql.context, f)
	return queryBuilder
}

func (mysql *MysqlDB) Update(table string) *UpdateBuilder {
	record := getRegisterRecord(table, mysql.Identify)
	updateBuilder := NewUpdateBuilder(mysql.Identify, mysql.masterDB(), mysql.context, record.(ActiveRecord))
	return updateBuilder
}

func (mysql *MysqlDB) Delete(table string) *DeleteBuilder {
	record := getRegisterRecord(table, mysql.Identify)
	deleteBuilder := NewDeleteBuilder(mysql.Identify, mysql.masterDB(), mysql.context, record.(ActiveRecord))
	return deleteBuilder
}

func (mysql *MysqlDB) Insert(table string) *InsertBuilder {
	record := getRegisterRecord(table, mysql.Identify)
	deleteBuilder := NewInsertBuilder(mysql.masterDB(), mysql.context, record.(ActiveRecord))
	return deleteBuilder
}

func (mysql *MysqlDB) BatchInsert(rs []ActiveRecord, columns ...string) int {
	return batchInsert(mysql.Identify, mysql.masterDB(), mysql.context, rs, columns...)
}

func (mysql *MysqlDB) SearchRecord(table string, mp map[string]string) *helper.RecordList {
	return searchRecord(mysql.Identify, mysql.slaveDB(), mysql.context, table, mp)
}

func (mysql *MysqlDB) SaveRecord(r ActiveRecord, columns ...string) int {
	return saveRecord(mysql.Identify, mysql.masterDB(), mysql.context, r, "", columns...)
}

func (mysql *MysqlDB) SaveRecordAndNotExist(r ActiveRecord, notExistSql string, columns ...string) int {
	return saveRecord(mysql.Identify, mysql.masterDB(), mysql.context, r, notExistSql, columns...)
}

func (mysql *MysqlDB) LoadRecord(r ActiveRecord) bool {
	return loadRecord(mysql.Identify, mysql.slaveDB(), mysql.context, r)
}

func (mysql *MysqlDB) DeleteRecord(r ActiveRecord) bool {
	return deleteRecord(mysql.Identify, mysql.masterDB(), mysql.context, r)
}

func (mysql *MysqlDB) Db() *sql.DB {
	return mysql.db
}

func (mysql *MysqlDB) BeginTransaction() *Transaction {
	tx, err := mysql.masterDB().Begin()
	checkerr(err)
	transaction := &Transaction{
		Identify: mysql.Identify,
		db:       tx,
		context: &Context{
			RequestId: mysql.context.RequestId,
			TxId:      helper.GenerateUUID(),
		},
		cache:    mysql.cache,
		useCache: mysql.useCache,
	}
	return transaction
}

////////////////////////事物////////////////////////////
type Transaction struct {
	Identify string
	db       *sql.Tx
	context  *Context
	cache    Cache
	useCache bool
}

var _ IDB = &Transaction{}

func (ts *Transaction) Query(recordName, sql string, params ...interface{}) *helper.RecordList {
	record := getRegisterRecord(recordName, ts.Identify)
	rows, err := ts.db.Query(sql, params...)
	logger.Debug("sql", ts.context.RequestId, sql)
	checkerr(err)
	stat := &QueryStat{rows: rows, record: record}
	return stat.FetchAll()
}

func (ts *Transaction) FindOneByPrimary(table string, id int) ActiveRecord {
	record := getRegisterRecord(table, ts.Identify)
	primaryField := getRecordPrimary(table, ts.Identify)
	queryBuilder := NewQueryBuilder(ts.Identify, ts.db, ts.context, record.(ActiveRecord))
	stat := queryBuilder.Select("*").Where(primaryField.Column, "=", id).Execute().Fetch()
	return stat
}

func (ts *Transaction) FindListByPrimary(table string, ids []int) *helper.RecordList {
	record := getRegisterRecord(table, ts.Identify)
	primaryField := getRecordPrimary(table, ts.Identify)
	queryBuilder := NewQueryBuilder(ts.Identify, ts.db, ts.context, record.(ActiveRecord))
	recordList := queryBuilder.Select("*").Where(primaryField.Column, "in", ids).Execute().FetchAll()
	return recordList
}

func (ts *Transaction) Find(table string) *QueryBuilder {
	record := getRegisterRecord(table, ts.Identify)
	queryBuilder := NewQueryBuilder(ts.Identify, ts.db, ts.context, record.(ActiveRecord))
	return queryBuilder
}

func (ts *Transaction) Finder(f Finder) *QueryBuilder {
	queryBuilder := NewQueryFinder(ts.Identify, ts.db, ts.context, f)
	return queryBuilder
}

func (ts *Transaction) Update(table string) *UpdateBuilder {
	record := getRegisterRecord(table, ts.Identify)
	updateBuilder := NewUpdateBuilder(ts.Identify, ts.db, ts.context, record.(ActiveRecord))
	return updateBuilder
}

func (ts *Transaction) Delete(table string) *DeleteBuilder {
	record := getRegisterRecord(table, ts.Identify)
	deleteBuilder := NewDeleteBuilder(ts.Identify, ts.db, ts.context, record.(ActiveRecord))
	return deleteBuilder
}

func (ts *Transaction) Insert(table string) *InsertBuilder {
	record := getRegisterRecord(table, ts.Identify)
	deleteBuilder := NewInsertBuilder(ts.db, ts.context, record.(ActiveRecord))
	return deleteBuilder
}

func (ts *Transaction) BatchInsert(rs []ActiveRecord, columns ...string) int {
	return batchInsert(ts.Identify, ts.db, ts.context, rs, columns...)
}

func (ts *Transaction) SearchRecord(table string, orderBy string, mp map[string]string) *helper.RecordList {
	return searchRecord(ts.Identify, ts.db, ts.context, table, mp)
}

func (ts *Transaction) SaveRecord(r ActiveRecord, columns ...string) int {
	return saveRecord(ts.Identify, ts.db, ts.context, r, "", columns...)
}

func (ts *Transaction) SaveRecordAndNotExist(r ActiveRecord, notExistSql string, columns ...string) int {
	return saveRecord(ts.Identify, ts.db, ts.context, r, notExistSql, columns...)
}

func (ts *Transaction) LoadRecord(r ActiveRecord) bool {
	return loadRecord(ts.Identify, ts.db, ts.context, r)
}

func (ts *Transaction) DeleteRecord(r ActiveRecord) bool {
	return deleteRecord(ts.Identify, ts.db, ts.context, r)
}

func (ts *Transaction) Commit() {
	checkerr(ts.db.Commit())
}

func (ts *Transaction) RollBack() {
	ts.db.Rollback()
}

func (ts *Transaction) CatchException() {
	if err := recover(); err != nil {
		ts.RollBack()
		panic(err)
	}
}

func (ts *Transaction) Tx() *sql.Tx {
	return ts.db
}

func searchRecord(identify string, db interface{}, c *Context, table string, mp map[string]string) *helper.RecordList {
	r := getRegisterRecord(table, identify)
	findBuilder := NewQueryBuilder(identify, db, c, r.(ActiveRecord)).Select("*")
	expressions := make([]*expression, 0, len(mp))
	orderBy := "`create_time`"
	orderTyp := "DESC"
	page := 1
	size := 10
	if _, ok := mp["page"]; ok {
		if p, err := strconv.Atoi(mp["page"]); err == nil && p > 0 {
			page = p
		}
		delete(mp, "page")
	}
	if _, ok := mp["size"]; ok {
		if s, err := strconv.Atoi(mp["size"]); err == nil && s > 0 {
			size = s
		}
		delete(mp, "size")
	}
	if _, ok := mp["orderBy"]; ok {
		t := getRecordField(table, mp["orderBy"], identify)
		if t != nil {
			orderBy = "`" + t.Column + "`"
		}
		delete(mp, "orderBy")
	}

	if _, ok := mp["orderType"]; ok {
		if strings.ToLower(mp["orderType"]) == "asc" || strings.ToLower(mp["orderType"]) == "desc" {
			orderTyp = strings.ToUpper(mp["orderType"])
		}
		delete(mp, "orderType")
	}

	for k, v := range mp {
		f := getRecordField(table, k, identify)
		if f == nil {
			continue
		}
		eps, err := getExpresstions(f.Column, f.Type, v)
		if err != nil {
			exception.Panic(exception.ExceptionInvalidParams, err)
		}
		expressions = append(expressions, eps...)
	}
	for _, ep := range expressions {
		findBuilder.Where(ep.key, ep.operate, ep.value)
	}
	recordList := findBuilder.OrderBy(orderBy + " " + orderTyp).Offset((page - 1) * size).Limit(size).Execute().FetchAll()
	return recordList
}

func getExpresstions(n, t, s string) ([]*expression, error) {
	result := make([]*expression, 0)
	if s == "" {
		return result, nil
	}
	if strings.Index(s, ",") < 0 {
		if strings.Index(s, "|") < 0 {
			if t == FieldTypeInt || s[len(s)-1:] != " " {
				e := checkType(t, s)
				if e != nil {
					return nil, e
				}
				result = append(result, &expression{
					key:     n,
					operate: "=",
					value:   s,
				})
			} else {
				vslice := strings.Split(s, " ")
				for _, v := range vslice {
					if v != "" {
						result = append(result, &expression{
							key:     n,
							operate: "like",
							value:   "%" + v + "%",
						})
					}
				}
			}

		} else {
			vslice := strings.Split(s, "|")
			for _, v := range vslice {
				e := checkType(t, v)
				if e != nil {
					return nil, e
				}
			}
			result = append(result, &expression{
				key:     n,
				operate: "in",
				value:   vslice,
			})
		}
	} else {
		vslice := strings.Split(s, ",")
		if len(vslice) != 2 {
			return nil, errors.New("too much value " + n + ":" + s)
		}
		for _, v := range vslice {
			if v != "" {
				e := checkType(t, v)
				if e != nil {
					return nil, e
				}
			}
		}
		if vslice[0] != "" {
			result = append(result, &expression{
				key:     n,
				operate: ">",
				value:   vslice[0],
			})
		}
		if vslice[1] != "" {
			result = append(result, &expression{
				key:     n,
				operate: "<=",
				value:   vslice[1],
			})
		}
	}
	return result, nil
}

func checkType(t, s string) error {
	if t == FieldTypeInt {
		_, e := strconv.ParseInt(s, 10, 32)
		if e != nil {
			return errors.New("give the wrong type:" + s)
		}
	}
	return nil
}

func saveRecord(identify string, db interface{}, c *Context, r ActiveRecord, notExistSql string, cs ...string) int {
	getRegisterRecord(r.Name(), identify)
	fields := getRecordFields(r.Name(), identify)
	columns := getFieldsOfColumn(fields, cs)
	if len(columns) == 0 {
		checkerr(errors.New("the columns to insert is empty"))
	}

	v := reflect.Indirect(reflect.ValueOf(r))
	kv := make(map[string]interface{})
	vs := make([]interface{}, 0, len(columns))
	cls := make([]string, 0, len(columns))
	for _, c := range columns {
		cls = append(cls, c.Column)
		kv[c.Column] = v.Field(c.Index).Interface()
		vs = append(vs, kv[c.Column])
	}
	primaryField := getRecordPrimary(r.Name(), identify)

	if v.Field(primaryField.Index).Int() <= 0 {
		insertBuilder := NewInsertBuilder(db, c, r.(ActiveRecord))
		stat := insertBuilder.Columns(cls...).Value(vs...).NotExist(notExistSql).Execute()
		v.Field(primaryField.Index).SetInt(int64(stat.lastInsertId))
		return stat.affectRows
	} else {
		updateBuilder := NewUpdateBuilder(identify, db, c, r.(ActiveRecord))
		for k, v := range kv {
			updateBuilder.Set(k, v)
		}
		return updateBuilder.Where(primaryField.Column, "=", v.Field(primaryField.Index).Int()).Execute()
	}
}

func loadRecord(identify string, db interface{}, c *Context, r ActiveRecord) bool {
	getRegisterRecord(r.Name(), identify)
	field, ok := registerRecords[identify][r.Name()]["primary"]
	if !ok {
		return false
	}
	v := reflect.Indirect(reflect.ValueOf(r))
	id := v.Field(field.(*RecordField).Index).Int()
	if id <= 0 {
		return false
	}
	queryBuilder := NewQueryBuilder(identify, db, c, r)
	t := queryBuilder.Select("*").Where(field.(*RecordField).Column, "=", id).Execute().Fetch(r)
	return t != nil
}

func deleteRecord(identify string, db interface{}, c *Context, r ActiveRecord) bool {
	getRegisterRecord(r.Name(), identify)
	field, ok := registerRecords[identify][r.Name()]["primary"]
	if !ok {
		return false
	}
	v := reflect.Indirect(reflect.ValueOf(r))
	id := v.Field(field.(*RecordField).Index).Int()
	if id <= 0 {
		return false
	}
	deleteBuilder := NewDeleteBuilder(identify, db, c, r)
	deleteBuilder.Where(field.(*RecordField).Column, "=", id).Execute()
	return true
}

func batchInsert(identify string, db interface{}, c *Context, rs []ActiveRecord, cs ...string) int {
	if len(rs) == 0 {
		return 0
	}
	getRegisterRecord(rs[0].Name(), identify)
	fields := getRecordFields(rs[0].Name(), identify)
	columns := getFieldsOfColumn(fields, cs)
	if len(columns) == 0 {
		checkerr(errors.New("the columns to insert is empty"))
	}
	cls := make([]string, 0, len(columns))
	for _, c := range columns {
		cls = append(cls, c.Column)
	}

	vss := make([][]interface{}, 0, len(rs))
	for _, r := range rs {
		v := reflect.Indirect(reflect.ValueOf(r))
		vs := make([]interface{}, 0, len(columns))
		for _, f := range columns {
			vs = append(vs, v.Field(f.Index).Interface())
		}
		vss = append(vss, vs)
	}

	insertBuilder := NewInsertBuilder(db, c, rs[0].(ActiveRecord))
	stat := insertBuilder.Columns(cls...).Values(vss...).Execute()
	return stat.affectRows

}

func getFieldsOfColumn(fields map[string]*RecordField, cs []string) []*RecordField {
	columns := make([]*RecordField, 0, len(cs))
	if len(cs) == 0 {
		for _, f := range fields {
			if f.Modify {
				columns = append(columns, f)
			}
		}
	} else {
		for _, c := range cs {
			if f, ok := fields[c]; ok && !f.IsPrimary {
				columns = append(columns, f)
			} else {
				for _, v := range fields {
					if !v.IsPrimary && strings.ToLower(v.Key) == strings.ToLower(c) {
						columns = append(columns, v)
						break
					}
				}
			}
		}
	}
	return columns
}

func RealEscapeString(value string) string {
	replace := map[string]string{"\\": "\\\\", "'": `\'`, "\\0": "\\\\0", "\n": "\\n", "\r": "\\r", `"`: `\"`, "\x1a": "\\Z"}

	for b, a := range replace {
		value = strings.Replace(value, b, a, -1)
	}

	return value
}

func checkerr(err error) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	exception.Panic(exception.ErrorMysqlPanic, err)
}
