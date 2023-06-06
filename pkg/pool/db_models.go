package pool

func DbModels() []interface{} {
	return []interface{}{&PoolBase{}, &PoolServiceBase{}, &PoolServiceAssociationBase{}, &OpLogPool{}}
}

func QueryDbModels() []interface{} {
	return []interface{}{&PoolBase{}, &OpLogPool{}, &PoolServiceBinding{}, &PoolServiceBase{}}
}
