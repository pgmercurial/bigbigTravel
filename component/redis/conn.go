package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"crypto/rand"

	"encoding/base64"

	"github.com/garyburd/redigo/redis"
	"bigbigTravel/component/exception"
	"bigbigTravel/component/helper"
	"bigbigTravel/component/logger"
)

type RedisConfig struct {
	Host        string
	Port        int64
	Db          int64
	MaxIdle     int64
	MaxActive   int64
	IdleTimeout int64
	Password    string
	LockTimeout int64
}

type RedisClient struct {
	pool *redis.Pool
}

type RedisConn struct {
	conn *redis.Conn
}

var config *RedisConfig
var instance *RedisClient
var unlockScript = redis.NewScript(1, `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end`)

func InitConfig(cfg *RedisConfig) {
	config = cfg
	instance = open()
}

func open() *RedisClient {
	if config == nil {
		return nil
	}
	pool := &redis.Pool{
		MaxActive:   int(config.MaxActive),
		MaxIdle:     int(config.MaxIdle),
		IdleTimeout: time.Duration(config.IdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
			if err != nil {
				return nil, err
			}
			if config.Password != "" {
				if _, err := conn.Do("AUTH", config.Password); err != nil {
					conn.Close()
					return nil, err
				}
			}
			conn.Do("select", config.Db)
			return conn, nil
		},
		Wait: true,
	}
	//set log
	go getCurrentRedisConnStatus(pool)
	return &RedisClient{
		pool: pool,
	}
}

func getCurrentRedisConnStatus(pool *redis.Pool) {
	timer := time.NewTimer(60 * time.Second) // 新建定时器
	for {
		select {
		case <-timer.C:
			timer.Reset(60 * time.Second) // 重新设置定时器
			var stat redis.PoolStats
			stat = pool.Stats()
			logger.Info("redis-conn", helper.GenerateUUID(), map[string]int{
				"当前活跃连接数": stat.ActiveCount - stat.IdleCount, // 使用的连接数 = 建立的连接数 - 空闲的连接数
				"当前空闲连接数": stat.IdleCount,
				"当前建立连接数": stat.ActiveCount,
				"最大建立连接数": pool.MaxActive,
				"最大空闲连接数": pool.MaxIdle,
			})
		}
	}
}

func GetInstance() *RedisConn {
	conn := instance.pool.Get()
	return &RedisConn{
		conn: &conn,
	}
}

func (c *RedisConn) Do(cmd string, args ...interface{}) (interface{}, error) {
	return (*c.conn).Do(cmd, args...)
}

func (c *RedisConn) Send(cmd string, args ...interface{}) error {
	return (*c.conn).Send(cmd, args...)
}

func (c *RedisConn) Close() error {
	return (*c.conn).Close()
}

func (c *RedisConn) Err() error {
	return (*c.conn).Err()
}

func (c *RedisConn) Del(key string) {
	_, err := c.Do("DEL", key)
	checkErr(err)
}

func (c *RedisConn) Set(key string, value interface{}, expire ...interface{}) {
	if len(expire) > 1 {
		checkErr(errors.New("too many args for redis set"))
	}
	if len(expire) == 0 {
		_, err := c.Do("SET", key, value)
		checkErr(err)
	} else {
		ex, ok := expire[0].(int)
		if !ok {
			checkErr(errors.New("set expire key failed, invalid expire"))
		}
		_, err := c.Do("SETEX", key, ex, value)
		checkErr(err)
	}

}

func (c *RedisConn) Lock(key string, expire int) (string, bool) {
	if expire <= 0 {
		return "", false
	}
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", false
	}
	value := base64.StdEncoding.EncodeToString(b)
	_, err = c.Do("SET", key, value, "ex", expire, "nx")
	if err != nil {
		logger.Warning("redis_lock", "", fmt.Sprintf("lock fail key %s value %v expire %v", key, value, expire))
		return "", false
	}
	return value, true
}

func (c *RedisConn) Unlock(key string, value string) {
	unlockScript.Do(*c.conn, key, value)
}

func (c *RedisConn) SetJson(key string, value interface{}, expire ...interface{}) {
	s, e := json.Marshal(value)
	if e != nil {
		checkErr(errors.New("set json value failed," + e.Error()))
	}
	c.Set(key, s, expire...)
}

func (c *RedisConn) Get(key string) string {
	reply, err := c.Do("GET", key)
	checkErr(err)
	if reply == nil {
		reply = []byte{}
	}
	return string(reply.([]byte))
}

func (c *RedisConn) GetJson(key string, obj interface{}) bool {
	reply, err := c.Do("GET", key)
	checkErr(err)
	if reply == nil {
		return false
	}
	v := reply.([]byte)
	err = json.Unmarshal(v, obj)
	if err == nil {
		return true
	} else {
		return false
	}
}

func (c *RedisConn) IncBy(key string, value int, expire ...interface{}) int {
	if len(expire) > 1 {
		checkErr(errors.New("too many args for redis incby"))
	}
	if len(expire) >= 0 {
		err := c.Send("INCRBY", key, value)
		checkErr(err)
	} else {
		err := c.Send("DECRBY", key, 0-value)
		checkErr(err)
	}
	if len(expire) == 1 {
		ex, ok := expire[0].(int)
		if !ok {
			checkErr(errors.New("set expire key failed, invalid expire"))
		}
		c.Send("EXPIRE", key, ex)
	}
	(*c.conn).Flush()
	r, err := (*c.conn).Receive()
	checkErr(err)
	return int(r.(int64))
}

func (c *RedisConn) MSet(kv map[string]interface{}, expire ...interface{}) (err error) {
	if len(expire) > 1 {
		checkErr(errors.New("too many args for redis mset"))
	}
	if len(expire) == 0 {
		args := make([]interface{}, 0, len(kv)*2)
		for k, v := range kv {
			args = append(args, k, v)
		}
		_, err = (*c.conn).Do("MSET", args...)
	} else {
		ex, ok := expire[0].(int)
		if !ok {
			checkErr(errors.New("set expire key failed, invalid expire"))
		}
		for k, v := range kv {
			(*c.conn).Send("SETEX", k, ex, v)
		}
		err = (*c.conn).Flush()
	}
	return
}

func (c *RedisConn) MGet(keys ...string) map[string]string {
	result := make(map[string]string)
	if len(keys) == 0 {
		return result
	}
	args := make([]interface{}, 0, len(keys))
	for _, k := range keys {
		args = append(args, k)
	}
	reply, err := (*c.conn).Do("MGET", args...)
	checkErr(err)
	resp := reply.([]interface{})

	for i, r := range resp {
		result[args[i].(string)] = string(r.([]byte))
	}
	return result
}

func (c *RedisConn) CreateHash(key string) *Hash {
	return &Hash{
		conn: c.conn,
		key:  key,
	}
}

func (c *RedisConn) CreateList(key string) *List {
	return &List{
		conn: c.conn,
		key:  key,
	}
}

func (c *RedisConn) CreateSet(key string) *Set {
	return &Set{
		conn: c.conn,
		key:  key,
	}
}

func (c *RedisConn) CreateSortSet(key string) *SortSet {
	return &SortSet{
		conn: c.conn,
		key:  key,
	}
}

func checkErr(err error) {
	if err == nil {
		return
	}
	fmt.Println(err.Error())
	exception.Panic(exception.ErrorRedisPanic, err)
}
