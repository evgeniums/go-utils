package pool

type PoolStore interface {
	Pool(id string) (Pool, error)
	PoolByName(name string) (Pool, error)
	Reload() error
}
