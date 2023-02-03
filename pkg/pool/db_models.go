package pool

func DbModels() []interface{} {
	return []interface{}{&PoolBase{}, &PoolServiceBase{}, &PostgresServer{}, &RestApiServer{}, &RedisServer{}}
}
