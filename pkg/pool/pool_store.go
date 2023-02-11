package pool

type PoolStore interface {
	Pool(id string) (Pool, error)
	Reload() error
}
