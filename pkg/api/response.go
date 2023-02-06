package api

import "github.com/evgeniums/go-backend-helpers/pkg/generic_error"

const (
	TargetSelf   = "self"
	TargetParent = "parent"
	TargetChild  = "parent"
)

type HateoasLink struct {
	Target     string `json:"target"`
	Operation  string `json:"operation"`
	HttpMethod string `json:"http_method"`
	Host       string `json:"host"`
	Path       string `json:"path"`
}

type HateoasLinks struct {
	Links []*HateoasLink `json:"links"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

func ResponseGenericError(responseError *ResponseError) generic_error.Error {
	if responseError == nil {
		return nil
	}
	e := generic_error.New(responseError.Code, responseError.Message)
	e.SetDetails(responseError.Details)
	return e
}
