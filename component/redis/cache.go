package redis

type Cache struct {
	redis 		*RedisConn
	expire		int
}

func (cache *Cache) Expire (second int) *Cache {
	cache.expire = second
	return cache
}

func (cache *Cache) Set (key string, value interface{})  {
	if !cache.checkRedis() {
		return
	}
	cache.redis.Set(key, value, cache.expire)
}

func (cache *Cache) SetJson (key string, value interface{})  {
	if !cache.checkRedis() {
		return
	}
	cache.redis.SetJson(key, value, cache.expire)
}

func (cache *Cache) Get (key string) string {
	if !cache.checkRedis() {
		return ""
	}
	return cache.redis.Get(key)
}

func (cache *Cache) GetJson (key string, obj interface{}) bool {
	if !cache.checkRedis() {
		return false
	}
	return cache.redis.GetJson(key, obj)
}

func (cache *Cache) SetMulti(kv map[string]interface{})  {
	if !cache.checkRedis() {
		return
	}
	cache.redis.MSet(kv, cache.expire)
}

func (cache *Cache) GetMulti(keys ...string) map[string]string {
	if !cache.checkRedis() {
		return make(map[string]string)
	}
	return cache.redis.MGet(keys...)
}

func (cache *Cache) checkRedis() bool {
	if cache.redis == nil || cache.redis.Err() != nil {
		return false
	}
	return true
}
