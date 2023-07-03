package cache

import "time"

type Lock interface {
	NotObtained() bool
	Release() error
}

type Locker interface {
	Lock(key string, ttl time.Duration) (Lock, error)
}
