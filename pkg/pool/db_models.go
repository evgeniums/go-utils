package pool

func DbModels() []interface{} {
	return []interface{}{&PoolBase{}, &PoolServiceBase{}, &PoolServiceBindingBase{}}
}
