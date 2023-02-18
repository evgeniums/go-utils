package pool

// TODO Implement pool store
type PoolStore interface {
	Pool(id string) (Pool, error)
	PoolByName(name string) (Pool, error)
}
