package api

import "github.com/evgeniums/go-backend-helpers/pkg/generic_error"

const (
	TargetSelf   = "self"
	TargetParent = "parent"
	TargetChild  = "parent"
)

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

type Response interface {
	WithHateoasLinks
}

type ResponseNoHateous struct {
	HateoasLinksStub
}

type ResponseHateous struct {
	HateoasLinksContainer
}

type ResponseCount struct {
	Count int64 `json:"count,omitempty"`
}

type ResponseExists struct {
	Response `json:"-"`
	Exists   bool `json:"exists"`
}
