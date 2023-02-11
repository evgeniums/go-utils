package admin

func DbModels() []interface{} {
	return []interface{}{&Admin{}, &AdminSession{}, &AdminSessionClient{}, &OpLogAdmin{}}
}
