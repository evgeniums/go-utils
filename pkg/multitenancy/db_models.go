package multitenancy

func DbModels() []interface{} {
	return []interface{}{&TenancyDb{}, &OpLogTenancy{}}
}
