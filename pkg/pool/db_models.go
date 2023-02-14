package pool

func DbModels() []interface{} {
	return []interface{}{&PoolBase{}, &PoolServiceBase{}, &PoolServiceAssociationBase{}, &OpLogPool{}}
}
