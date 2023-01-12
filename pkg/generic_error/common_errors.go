package generic_error

import "net/http"

const (
	ErrorCodeUnknown                    string = "unknown_error"
	ErrorCodeInternalServerError        string = "internal_server_error"
	ErrorCodeFormat                     string = "invalid_format"
	ErrorCodeFieldValue                 string = "invalid_field_value"
	ErrorCodeValidation                 string = "validation_failed"
	ErrorCodeForbidden                  string = "forbidden"
	ErrorCodeUnauthorized               string = "unauthorized"
	ErrorCodeNotFound                   string = "resource_not_found"
	ErrorCodeExternalServiceUnavailable string = "external_service_unavailable"
	ErrorCodeExternalServiceError       string = "external_service_error"
)

var CommonErrorDescriptions = map[string]string{
	ErrorCodeUnknown:                    "Unknown error",
	ErrorCodeInternalServerError:        "Internal server error",
	ErrorCodeFormat:                     "Invalid format of request",
	ErrorCodeFieldValue:                 "Invalid value of request field",
	ErrorCodeValidation:                 "Validation failed",
	ErrorCodeForbidden:                  "Access denied",
	ErrorCodeUnauthorized:               "Request is not authorized",
	ErrorCodeNotFound:                   "Resource is not found",
	ErrorCodeExternalServiceUnavailable: "External service is temporarily unavailable",
	ErrorCodeExternalServiceError:       "External service reported error",
}

var CommonErrorHttpCodes = map[string]int{
	ErrorCodeUnknown:                    http.StatusInternalServerError,
	ErrorCodeInternalServerError:        http.StatusInternalServerError,
	ErrorCodeForbidden:                  http.StatusForbidden,
	ErrorCodeUnauthorized:               http.StatusUnauthorized,
	ErrorCodeNotFound:                   http.StatusNotFound,
	ErrorCodeExternalServiceUnavailable: http.StatusInternalServerError,
	ErrorCodeExternalServiceError:       http.StatusInternalServerError,
}
