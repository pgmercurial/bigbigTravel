package mysql

import (
	"sync"
	"testing"
)

const RecordNameGolibTest = "golib_test"

func init() {
	var r = &GolibTest{}
	RegisterRecord(r.Name(), r)
}

type GolibTest struct {
	GolibTestzId int    `json:"golibTestzId" form:"golibTestzId" column:"golib_testz_id" primary:"true" modify:"false"`
	UserId       int    `json:"userId" form:"userId" column:"user_id" modify:"true"`
	TestTimes    int    `form:"testTimes" column:"test_times" modify:"true" json:"testTimes"`
	CourseId     int    `json:"courseId" form:"courseId" column:"course_id" modify:"true"`
	TermId       int    `json:"termId" form:"termId" column:"term_id" modify:"true"`
	CreateTime   string `json:"createTime" form:"createTime" column:"create_time" modify:"false"`
	UpdateTime   string `form:"updateTime" column:"update_time" modify:"false" json:"updateTime"`
}

func (r *GolibTest) Name() string {
	return RecordNameGolibTest
}

func TestInsertNotExist(t *testing.T) {
	InitConfig(&MysqlConfig{
		Host:            "127.0.0.1",
		Port:            12345,
		UserName:        "test",
		PassWord:        "test",
		DBName:          "aiclass",
		MaxConnLifeTime: 3600,
		MaxIdleConns:    32,
		MaxOpenConns:    200,
	})

	db := GetInstance(false)
	resp := db.SaveRecord(&GolibTest{UserId: 1, TestTimes: 1})
	if resp != 1 {
		panic("save golib test records return not 1")
	}
	resp = db.SaveRecordAndNotExist(&GolibTest{UserId: 1, TestTimes: 2}, "select user_id from golib_test where user_id = 1")
	if resp != 0 {
		panic("save not exist golib test records return not 0")
	}
	resp = db.SaveRecordAndNotExist(&GolibTest{UserId: 2, TestTimes: 2}, "select user_id from golib_test where user_id = 2")
	if resp != 1 {
		panic("save not exist golib test records return not 1")
	}
	resp = db.SaveRecordAndNotExist(&GolibTest{UserId: 3, TestTimes: 2, CourseId: 111, TermId: 222}, "select user_id from golib_test where user_id = 3")
	if resp != 1 {
		panic("save not exist user 3 golib test records return not 1")
	}

	ts := db.BeginTransaction()
	defer ts.CatchException()
	resp = ts.SaveRecord(&GolibTest{UserId: 11, TestTimes: 1})
	if resp != 1 {
		panic("save golib test records return not 1")
	}
	resp = ts.SaveRecordAndNotExist(&GolibTest{UserId: 11, TestTimes: 2}, "select user_id from golib_test where user_id = 11")
	if resp != 0 {
		panic("save not exist golib test records return not 0")
	}
	resp = ts.SaveRecordAndNotExist(&GolibTest{UserId: 22, TestTimes: 2}, "select user_id from golib_test where user_id = 22")
	if resp != 1 {
		panic("save not exist golib test records return not 1")
	}
	resp = ts.SaveRecordAndNotExist(&GolibTest{UserId: 33, TestTimes: 2, CourseId: 111, TermId: 222}, "select user_id from golib_test where user_id = 33")
	if resp != 1 {
		panic("save not exist user 3 golib test records return not 1")
	}
	ts.Commit()
}

func TestInsertNotExistCurrent(t *testing.T) {
	InitConfig(&MysqlConfig{
		Host:            "127.0.0.1",
		Port:            12345,
		UserName:        "test",
		PassWord:        "test",
		DBName:          "aiclass",
		MaxConnLifeTime: 3600,
		MaxIdleConns:    32,
		MaxOpenConns:    200,
	})

	db := GetInstance(false)
	wg := new(sync.WaitGroup)
	record := &GolibTest{CourseId: 11, TermId: 22, UserId: 33, TestTimes: 44}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()
			ts := db.BeginTransaction()
			defer ts.CatchException()
			ts.SaveRecordAndNotExist(record, "select 1 from golib_test where course_id = 11")
			//ts.SaveRecord(record)
			ts.Commit()
		}()
	}
	wg.Wait()
}

func TestInsertPrepare(t *testing.T) {
	InitConfig(&MysqlConfig{
		Host:            "127.0.0.1",
		Port:            12345,
		UserName:        "test",
		PassWord:        "test",
		DBName:          "aiclass",
		MaxConnLifeTime: 3600,
		MaxIdleConns:    32,
		MaxOpenConns:    200,
	})

	db := GetInstance(false)
	stmt, _ := db.db.Prepare("INSERT IGNORE INTO `golib_test` (user_id, course_id, term_id) SELECT ?,?,? FROM dual WHERE not exists (select 1 from `golib_test` where course_id = 1111)")
	stmt.Exec(11, 22, 33)
}
