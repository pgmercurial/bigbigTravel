package redis

import (
	"github.com/garyburd/redigo/redis"
	"errors"
	"encoding/json"
)

type List struct {
	conn	*redis.Conn
	key		string
}

func (l *List) Expire(seconds int)  {
	_, err := (*l.conn).Do("EXPIRE", l.key, seconds)
	checkErr(err)
}

func (l *List) Exists() bool {
	reply, err := (*l.conn).Do("EXISTS", l.key)
	checkErr(err)
	return reply.(int64) == 1
}

func (l *List) Len() int {
	reply, err := (*l.conn).Do("LLEN", l.key)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) LPush(value interface{}) int {
	reply, err := (*l.conn).Do("LPUSH", l.key, value)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) LPushJson(value interface{}) int {
	str, err := json.Marshal(value)
	checkErr(err)
	reply, err := (*l.conn).Do("LPUSH", l.key, str)
	checkErr(err)
	return int(reply.(int64))
}

//PUSH if key exists
func (l *List) LPushX(value interface{}) int {
	reply, err := (*l.conn).Do("LPUSHX", l.key, value)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) LPushJsonX(value interface{}) int {
	str, err := json.Marshal(value)
	checkErr(err)
	reply, err := (*l.conn).Do("LPUSHX", l.key, str)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) LPop() (string, bool) {
	reply, err := (*l.conn).Do("LPOP", l.key)
	checkErr(err)
	if reply == nil{
		return "", false
	}
	return string(reply.([]byte)), true
}

func (l *List) RPush(value interface{}) int {
	reply, err := (*l.conn).Do("RPUSH", l.key, value)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) RPushJson(value interface{}) int {
	str, err := json.Marshal(value)
	checkErr(err)
	reply, err := (*l.conn).Do("RPUSH", l.key, str)
	checkErr(err)
	return int(reply.(int64))
}

//PUSH if key exists
func (l *List) RPushX(value interface{}) int {
	reply, err := (*l.conn).Do("RPUSHX", l.key, value)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) RPushJsonX(value interface{}) int {
	str, err := json.Marshal(value)
	checkErr(err)
	reply, err := (*l.conn).Do("RPUSHX", l.key, str)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) RPop() (string, bool) {
	reply, err := (*l.conn).Do("RPOP", l.key)
	checkErr(err)
	if reply == nil{
		return "", false
	}
	return string(reply.([]byte)), true
}

//not found "before" value return -1, success return new len of list, return 0 when key not exists or empty list
func (l *List) InsertBefore(value interface{}, before interface{}) int {
	reply, err := (*l.conn).Do("LINSERT", l.key, "BEFORE", before, value)
	checkErr(err)
	return int(reply.(int64))
}

//not found "after" value return -1, success return new len of list, return 0 when key not exists or empty list
func (l *List) InsertAfter(value interface{}, after interface{}) int {
	reply, err := (*l.conn).Do("LINSERT", l.key, "AFTER", after, value)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) Index(index int) (string, bool) {
	reply, err := (*l.conn).Do("LINDEX", l.key, index)
	checkErr(err)
	if reply == nil {
		return "", false
	}else {
		return string(reply.([]byte)), true
	}
}

func (l *List) Range(s , e int) []string {
	result := make([]string, 0)
	reply, err := (*l.conn).Do("LRANGE", l.key, s, e)
	checkErr(err)
	if reply == nil {
		return result
	}
	r,o := reply.([]interface{})
	if !o {
		checkErr(errors.New("list range failed on replay assert"))
	}
	for _,v := range r {
		result = append(result, string(v.([]byte)))
	}
	return result
}

func (l *List) Remove(count int, val interface{}) int {
	reply, err := (*l.conn).Do("LREM", l.key, count, val)
	checkErr(err)
	return int(reply.(int64))
}

func (l *List) Set(index int, val interface{}) {
	_, err := (*l.conn).Do("LSET", l.key, index, val)
	checkErr(err)
}

//slice from list as the new list
func (l *List) Trim(s , e int) {
	_, err := (*l.conn).Do("LTRIM", l.key, s, e)
	checkErr(err)
}

func (l *List) MoveTo(key string) string {
	reply, err := (*l.conn).Do("RPOPLPUSH", l.key, key)
	checkErr(err)
	return string(reply.([]byte))
}


