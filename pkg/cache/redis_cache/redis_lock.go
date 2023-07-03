package redis_cache

import (
	"time"

	"github.com/bsm/redislock"
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
)

type RedisLocker struct {
	*RedisCache
	locker *redislock.Client
}

type RedisLock struct {
	locker      *RedisLocker
	lock        *redislock.Lock
	notObtained bool
}

func (r *RedisLock) Release() error {
	err := r.lock.Release(r.locker.Context())
	if err != nil {
		if err == redislock.ErrLockNotHeld {
			return nil
		}
		return err
	}
	return nil
}

func (r *RedisLock) NotObtained() bool {
	return r.notObtained
}

func NewLocker(redisCache *RedisCache) *RedisLocker {
	l := &RedisLocker{}
	l.locker = redislock.New(redisCache.NativeHandler())
	return l
}

func (r *RedisLocker) Lock(key string, ttl time.Duration) (cache.Lock, error) {

	var err error

	lock := &RedisLock{locker: r}
	lock.lock, err = r.locker.Obtain(r.Context(), key, ttl, nil)
	if err != nil {

		if err == redislock.ErrNotObtained {
			lock.lock = nil
			lock.notObtained = true
			return lock, nil
		}

		return nil, err
	}

	return lock, nil
}
