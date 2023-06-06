package admin

func DbModels() []interface{} {
	return []interface{}{&Admin{}, &AdminSession{}, &AdminSessionClient{}, &OpLogAdmin{}}
}

func QueryDbModels() []interface{} {
	return DbModels()
}
