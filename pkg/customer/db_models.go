package customer

func DbModels() []interface{} {
	return []interface{}{&Customer{}, &CustomerSession{}, &CustomerSessionClient{}, &OpLogCustomer{}}
}
