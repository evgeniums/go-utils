package validator

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
)

type Validator interface {
	Validate(s interface{}) error
	ValidateValue(value interface{}, rules string) error
	ValidatePartial(s interface{}, fields ...string) error
}

type ValidationError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Err     error  `json:"error"`
}

func (e *ValidationError) Error() string {
	if e.Message == "" {
		if e.Err == nil {
			return ""
		}
		return e.Err.Error()
	}
	if e.Field != "" {
		return fmt.Sprintf("validation failed on field \"%s\": %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed: %s", e.Message)
}

func (e *ValidationError) GenericError() generic_error.Error {
	err := generic_error.New(generic_error.ErrorCodeFormat, e.Message)
	err.SetDetails(e.Field)
	return err
}
