package redis

import (
	"github.com/garyburd/redigo/redis"
	"errors"
)

type Set struct {
	conn	*redis.Conn
	key		string
}

func (s *Set) Expire(seconds int)  {
	_, err := (*s.conn).Do("EXPIRE", s.key, seconds)
	checkErr(err)
}

func (s *Set) Exists() bool {
	reply, err := (*s.conn).Do("EXISTS", s.key)
	checkErr(err)
	return reply.(int64) == 1
}

func (s *Set) Add (value ...interface{}) int {
	if len(value) == 0 {
		return 0
	}
	args := []interface{}{s.key}
	for _,val := range value {
		args = append(args, val)
	}
	reply,err := (*s.conn).Do("SADD", args...)
	checkErr(err)
	return int(reply.(int64))
}

func (s *Set) Count () int {
	reply,err := (*s.conn).Do("SCARD", s.key)
	checkErr(err)
	return int(reply.(int64))
}

func (s *Set) Diff (key ...interface{}) []string {
	result := make([]string, 0)
	args := []interface{}{s.key}
	for _,val := range key {
		args = append(args, val)
	}
	reply,err := (*s.conn).Do("SDIFF", args...)
	checkErr(err)
	r, ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("set diff failed on reply assert"))
	}
	for _,v := range r {
		result = append(result, string(v.([]byte)))
	}
	return result
}

func (s *Set) DiffStore(dest string, diffKey ...string) int {
	args := []interface{}{dest, s.key}
	for _,val := range diffKey {
		args = append(args, val)
	}
	reply,err := (*s.conn).Do("SDIFFSTORE", args...)
	checkErr(err)
	return int(reply.(int64))
}

func (s *Set) Inter (key ...interface{}) []string {
	result := make([]string, 0)
	args := []interface{}{s.key}
	for _,val := range key {
		args = append(args, val)
	}
	reply,err := (*s.conn).Do("SINTER", args...)
	checkErr(err)
	r, ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("set inter failed on reply assert"))
	}
	for _,v := range r {
		result = append(result, string(v.([]byte)))
	}
	return result
}

func (s *Set) InterStore (dest string, interKey ...string) int {
	args := []interface{}{dest, s.key}
	for _,val := range interKey {
		args = append(args, val)
	}
	reply,err := (*s.conn).Do("SINTERSTORE", args...)
	checkErr(err)
	return int(reply.(int64))
}

func (s *Set) Union (key ...interface{}) []string {
	result := make([]string, 0)
	args := []interface{}{s.key}
	for _,val := range key {
		args = append(args, val)
	}
	reply,err := (*s.conn).Do("SUNION", args...)
	checkErr(err)
	r, ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("set union failed on reply assert"))
	}
	for _,v := range r {
		result = append(result, string(v.([]byte)))
	}
	return result
}

func (s *Set) UnionStore (dest string, unionKey ...string) int {
	args := []interface{}{dest, s.key}
	for _,val := range unionKey {
		args = append(args, val)
	}
	reply,err := (*s.conn).Do("SUNIONSTORE", args...)
	checkErr(err)
	return int(reply.(int64))
}

func (s *Set) IsMember (value interface{}) bool {
	reply,err := (*s.conn).Do("SISMEMBER", s.key, value)
	checkErr(err)
	return reply.(int64) == 1
}

func (s *Set) Members () []string {
	reply,err := (*s.conn).Do("SMEMBERS", s.key)
	checkErr(err)
	r, ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("set members failed on reply assert"))
	}
	result := make([]string, 0, len(r))
	for _,v := range r {
		result = append(result, string(v.([]byte)))
	}
	return result
}

func (s *Set) Move (destKey string, member interface{}) bool {
	reply,err := (*s.conn).Do("SMOVE", s.key, destKey, member)
	checkErr(err)
	return reply.(int64) == 1
}

func (s *Set) Pop () string {
	reply,err := (*s.conn).Do("SPOP", s.key)
	checkErr(err)
	return string(reply.([]byte))
}

func (s *Set) RandMember(count int) []string {
	reply,err := (*s.conn).Do("SRANDMEMBER", s.key, count)
	checkErr(err)
	r, ok := reply.([]interface{})
	if !ok {
		checkErr(errors.New("set rand member failed on reply assert"))
	}
	result := make([]string, 0, len(r))
	for _,v := range r {
		result = append(result, string(v.([]byte)))
	}
	return result
}

func (s *Set) Remove (value ...interface{}) int {
	args := []interface{}{s.key}
	for _, val := range  value {
		args = append(args, val)
	}
	reply,err := (*s.conn).Do("SREM", args...)
	checkErr(err)
	return int(reply.(int64))
}







