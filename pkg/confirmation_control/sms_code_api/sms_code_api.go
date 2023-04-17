package sms_code_api

import "github.com/evgeniums/go-backend-helpers/pkg/api"

const ServiceName string = "sms"

const OperationResource string = "operation"

func CheckSms() api.Operation {
	return api.Post("check_sms")
}

func PrepareCheckSms() api.Operation {
	return api.Get("prepare_check_sms")
}

func PrepareOperation() api.Operation {
	return api.Post("prepare_operation")
}

type Operation struct {
	Id        string `json:"id" validate:"required,id" vmessage:"Operation ID must be specified"`
	Phone     string `json:"phone" validate:"required,phone" vmessage:"Invalid phone format"`
	FailedUrl string `json:"failed_url" validate:"required,url" vmessage:"Invalid format of failed URL"`
}

type PrepareOperationResponse struct {
	api.ResponseStub
	Url string `json:"url"`
}

type PrepareCheckSmsResponse struct {
	api.ResponseStub
	FailedUrl string `json:"failed_url"`
}
