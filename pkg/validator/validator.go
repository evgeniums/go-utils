package validator

import "fmt"

type Validator interface {
	Validate(s interface{}) error
	ValidateValue(value interface{}, rules string) error
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
