package redis

import (
	"github.com/garyburd/redigo/redis"
	"fmt"
	"strconv"
)

type SortSet struct {
	conn 	*redis.Conn
	key		string
}

type SortItem struct {
	Score	int
	Value	string
}

func (ss *SortSet) Expire(seconds int)  {
	_, err := (*ss.conn).Do("EXPIRE", ss.key, seconds)
	checkErr(err)
}

func (ss *SortSet) Exists() bool {
	reply, err := (*ss.conn).Do("EXISTS", ss.key)
	checkErr(err)
	return reply.(int64) == 1
}

func (ss *SortSet) Add (score int, value interface{}) bool {
	reply,err := (*ss.conn).Do("ZADD", ss.key, score, value)
	checkErr(err)
	return reply.(int64) == 1
}

func (ss *SortSet) AllCount () int {
	reply,err := (*ss.conn).Do("ZCARD", ss.key)
	checkErr(err)
	return int(reply.(int64))
}

func (ss *SortSet) CountByScore (min, max int) int {
	reply,err := (*ss.conn).Do("ZCOUNT", ss.key, min, max)
	checkErr(err)
	return int(reply.(int64))
}

//返回新的score值
func (ss *SortSet) IncScore (score int, value interface{}) int {
	reply,err := (*ss.conn).Do("ZINCRBY", ss.key, score, value)
	checkErr(err)
	if reply == nil {
		return -1
	}
	score, e := strconv.Atoi(string(reply.([]byte)))
	checkErr(e)
	return int(score)
}

func (ss *SortSet) GetScore (value interface{}) int {
	reply,err := (*ss.conn).Do("ZSCORE", ss.key, value)
	checkErr(err)
	if reply == nil {
		return -1
	}
	score, e := strconv.Atoi(string(reply.([]byte)))
	checkErr(e)
	return int(score)
}



func (ss *SortSet) AscRangeByIndex (s,  e int) []*SortItem {
	reply,err := (*ss.conn).Do("ZRANGE", ss.key, s, e, "WITHSCORES")
	checkErr(err)
	r := reply.([]interface{})
	result := make([]*SortItem, 0, len(r))
	for i:=0; i<len(r); i+=2 {
		score, e := strconv.Atoi(string(r[i+1].([]byte)))
		checkErr(e)
		item := &SortItem{
			Score: score,
			Value: string(r[i].([]byte)),
		}
		result = append(result, item)
	}
	return result
}

func (ss *SortSet) DescRangeByIndex (s,  e int) []*SortItem {
	reply,err := (*ss.conn).Do("ZREVRANGE", ss.key, s, e, "WITHSCORES")
	checkErr(err)
	r := reply.([]interface{})
	result := make([]*SortItem, 0, len(r))
	for i:=0; i<len(r); i+=2 {
		score, e := strconv.Atoi(string(r[i+1].([]byte)))
		checkErr(e)
		item := &SortItem{
			Score: score,
			Value: string(r[i].([]byte)),
		}
		result = append(result, item)
	}
	return result
}

func (ss *SortSet) AscRankOf(value interface{}) int {
	reply,err := (*ss.conn).Do("ZRANK", ss.key, value)
	checkErr(err)
	if reply == nil{
		return -1
	}
	return int(reply.(int64))
}

func (ss *SortSet) DescRankOf(value interface{}) int {
	reply,err := (*ss.conn).Do("ZREVRANK", ss.key, value)
	checkErr(err)
	if reply == nil {
		return 0
	}
	return int(reply.(int64))
}

func (ss *SortSet) RemoveByValue(value ...interface{}) int {
	if len(value) == 0 {
		return 0
	}
	args := []interface{}{ss.key}
	for _,val := range value {
		args = append(args, val)
	}
	reply,err := (*ss.conn).Do("ZREM", args...)
	checkErr(err)
	return int(reply.(int64))
}

func (ss *SortSet) RemoveByRank(s, e int) int {
	if e < s {
		return 0
	}
	reply,err := (*ss.conn).Do("ZREMRANGEBYRANK", ss.key, s, e)
	checkErr(err)
	return int(reply.(int64))
}

func (ss *SortSet) RemoveByScore(s, e int) int {
	reply,err := (*ss.conn).Do("ZREMRANGEBYSCORE", ss.key, s, e)
	checkErr(err)
	return int(reply.(int64))
}

func (ss *SortSet) FindByScore() *SortFinder {
	return &SortFinder{
		conn:ss.conn,
		key:ss.key,
		offset:0,
		min: "-inf",
		max: "+inf",
	}
}


type SortFinder struct {
	conn	*redis.Conn
	key 	string
	min		interface{}
	max		interface{}
	offset 	int
	limit	interface{}
	order	string
}

func (sf *SortFinder) MinScore(s int, equal bool) *SortFinder {
	min := fmt.Sprint(s)
	if !equal {
		min = "("+min
	}
	sf.min = min
	return sf
}

func (sf *SortFinder) MaxScore(s int, equal bool) *SortFinder {
	max := fmt.Sprint(s)
	if !equal {
		max = "("+max
	}
	sf.max = max
	return sf
}

func (sf *SortFinder) Offset(offset int) *SortFinder {
	sf.offset = offset
	return sf
}

func (sf *SortFinder) Limit(limit int) *SortFinder {
	sf.limit = limit
	return sf
}

func (sf *SortFinder) OrderAsc() *SortFinder {
	sf.order = "asc"
	return sf
}

func (sf *SortFinder) OrderDesc() *SortFinder {
	sf.order = "desc"
	return sf
}

func (sf *SortFinder) Execute() []*SortItem {
	args := []interface{}{sf.key}
	var cmd string
	if sf.order == "asc" {
		args = append(args, sf.min, sf.max)
		cmd = "ZRANGEBYSCORE"
	}else {
		args = append(args, sf.max, sf.min)
		cmd = "ZREVRANGEBYSCORE"
	}
	args = append(args, "WITHSCORES")
	if sf.limit != nil {
		args = append(args, "LIMIT", sf.offset, sf.limit)
	}
	fmt.Println(cmd, args)
	reply,err := (*sf.conn).Do(cmd, args...)
	checkErr(err)
	r := reply.([]interface{})
	result := make([]*SortItem, 0, len(r))
	for i:=0; i<len(r); i+=2 {
		score, e := strconv.Atoi(string(r[i+1].([]byte)))
		checkErr(e)
		item := &SortItem{
			Score: score,
			Value: string(r[i].([]byte)),
		}
		result = append(result, item)
	}
	return result
}








