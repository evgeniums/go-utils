package redis_cache

import (
	"time"

	"github.com/evgeniums/go-backend-helpers/pkg/pubsub/pubsub_providers/pubsub_redis"
	"github.com/jellydator/ttlcache/v3"
	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	pubsub_redis.RedisClient
}

func New() *RedisCache {
	r := &RedisCache{}
	return r
}

func (r *RedisCache) Set(key string, value string, ttlSeconds ...int) error {

	ttl := ttlcache.NoTTL
	if len(ttlSeconds) > 0 {
		ttl = time.Second * time.Duration(ttlSeconds[0])
	}

	err := r.NativeHandler().SetEx(r.Context(), key, value, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisCache) Get(key string, value *string) (bool, error) {

	var err error
	*value, err = r.NativeHandler().Get(r.Context(), key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *RedisCache) Unset(key string) error {
	return r.NativeHandler().Del(r.Context(), key).Err()
}

func (r *RedisCache) Clear() error {
	return r.NativeHandler().FlushAll(r.Context()).Err()
}

func (r *RedisCache) Touch(key string) error {
	return r.NativeHandler().Touch(r.Context(), key).Err()
}

func (r *RedisCache) Start() {
}

func (r *RedisCache) Stop() {
	r.Shutdown(r.Context())
}

func (r *RedisCache) Keys() ([]string, error) {

	keys, err := r.NativeHandler().Keys(r.Context(), "*").Result()
	if err != nil {
		return nil, err
	}

	return keys, nil
}
