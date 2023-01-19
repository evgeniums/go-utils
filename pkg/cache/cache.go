package cache

type GenericCache[T any] interface {
	Set(key string, value T, ttlSeconds ...int) error
	Get(key string) (T, bool, error)
	Unset(key string) error
	Touch(key string) error
	Keys() ([]string, error)
	Clear() error
}

type Cache = GenericCache[interface{}]
