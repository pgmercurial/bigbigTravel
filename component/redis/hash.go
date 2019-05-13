package redis

import (
	"github.com/garyburd/redigo/redis"
	"errors"
	"strconv"
)

type Hash struct {
	conn		*redis.Conn
	key			string
}


func (h *Hash) Expire(seconds int)  {
	_, err := (*h.conn).Do("EXPIRE", h.key, seconds)
	checkErr(err)
}

func (h *Hash) Exists() bool {
	reply, err := (*h.conn).Do("EXISTS", h.key)
	checkErr(err)
	if reply == nil {
		return false
	}
	return reply.(int64) == 1
}

func (h *Hash) Get(field string) string {
	reply, err := (*h.conn).Do("HGET", h.key, field)
	checkErr(err)
	if reply == nil {
		return ""
	}
	return string(reply.([]byte))
}

func (h *Hash) Set(field string, value interface{})  {
	_, err := (*h.conn).Do("HSET", h.key, field, value)
	checkErr(err)
}

func (h *Hash) SetNx(field string, value interface{})  {
	_, err := (*h.conn).Do("HSETNX", h.key, field, value)
	checkErr(err)
}

func (h *Hash) Del(field ...string) int {
	if len(field) == 0 {
		return 0
	}
	args := make([]interface{}, 0, len(field) + 1)
	args = append(args, h.key)
	for _,v := range field {
		args = append(args, v)
	}
	reply, err := (*h.conn).Do("HDEL", args...)
	checkErr(err)
	if reply == nil {
		return 0
	}
	return int(reply.(int64))
}

func (h *Hash) Len() int {
	reply, err := (*h.conn).Do("HLEN", h.key)
	checkErr(err)
	if reply == nil {
		return 0
	}
	return int(reply.(int64))
}

func (h *Hash) GetAll() map[string]string {
	result := make(map[string]string)
	reply, err := (*h.conn).Do("HGETALL", h.key)
	checkErr(err)
	if reply == nil {
		return map[string]string{}
	}
	r, ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("hash getall failed on reply assert"))
	}
	for i:=0; i< len(r); i+=2  {
		result[string(r[i].([]byte))] = string(r[i+1].([]byte))
	}
	return result
}

func (h *Hash) FieldExists(field string) bool {
	reply, err := (*h.conn).Do("HEXISTS", h.key, field)
	checkErr(err)
	if reply == nil {
		return false
	}
	return reply.(int64) == 1
}

func (h *Hash) MGet(field ...string) map[string]string {
	result := make(map[string]string)
	if len(field) == 0 {
		return result
	}
	args := []interface{}{h.key}
	for _,f := range field {
		args = append(args, f)
	}
	reply, err := (*h.conn).Do("HMGET", args...)
	if reply == nil {
		return map[string]string{}
	}
	checkErr(err)
	r, ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("hash mget failed on reply assert"))
	}
	for i,v := range r{
		result[field[i]] = string(v.([]byte))
	}
	return result
}

func (h *Hash) MSet(kv map[string]interface{}){
	if len(kv) == 0 {
		return
	}
	args := make([]interface{},0,len(kv)*2+1)
	args = append(args, h.key)
	for k,v := range kv {
		args = append(args, k, v)
	}
	_, err := (*h.conn).Do("HMSET", args...)
	checkErr(err)
}

func (h *Hash) Values() []string {
	reply,err := (*h.conn).Do("HVALS", h.key)
	checkErr(err)
	if reply == nil {
		return []string{}
	}
	r,ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("hash values failed on reply assert"))
	}
	result := make([]string, 0, len(r))
	for _,v := range r {
		result = append(result, string(v.([]byte)))
	}
	return result
}

func (h *Hash) Keys() []string {
	reply,err := (*h.conn).Do("HKEYS", h.key)
	checkErr(err)
	if reply == nil {
		return []string{}
	}
	r,ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("hash keys failed on reply assert"))
	}
	result := make([]string, 0, len(r))
	for _,v := range r {
		result = append(result, string(v.([]byte)))
	}
	return result
}

func (h *Hash) IncByInt(field string, val int) int {
	reply, err := (*h.conn).Do("HINCRBY", h.key, field, val)
	checkErr(err)
	if reply == nil {
		return 0
	}
	return int(reply.(int64))
}

func (h *Hash) IncByFloat(field string, val float64) float64 {
	reply, err := (*h.conn).Do("HINCRBYFLOAT", h.key, field, val)
	checkErr(err)
	if reply == nil {
		return 0
	}
	f,e := strconv.ParseFloat(string(reply.([]byte)), 64)
	if e != nil {
		checkErr(errors.New("hash keys float incr failed on reply assert"))
	}
	return f
}

func (h *Hash) Drop()  {
	_,err := (*h.conn).Do("DEL", h.key)
	checkErr(err)
}





