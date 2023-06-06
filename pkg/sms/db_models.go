package sms

func DbModels() []interface{} {
	return []interface{}{&SmsMessage{}}
}

func QueryDbModels() []interface{} {
	return []interface{}{&SmsMessage{}}
}
