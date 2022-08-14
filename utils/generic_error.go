package utils

import "fmt"

type Error struct {
	Code    string
	Message string
	Details string
	Native  error
}

func NewError(code string, message ...string) *Error {
	e := &Error{Code: code}
	if len(message) > 0 {
		e.Message = message[0]
	}
	return e
}

func NewNativeError(code string, message string, err error) *Error {
	e := &Error{Code: code, Message: message, Native: err}
	return e
}

func NewSystemError(code string, err error) *Error {
	return NewError(code, err.Error())
}

func NewNativeErrorWithMessage(err error, message ...string) *Error {
	e := &Error{Native: err, Code: "unknown_error"}
	if len(message) > 0 {
		e.Message = fmt.Sprintf("%v: %v", err, message[0])
	} else {
		e.Message = fmt.Sprintf("%v", err)
	}
	return e
}

func NewErrorFromMessage(message string) *Error {
	e := &Error{Code: "unknown_error", Message: message}
	return e
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Value() string {
	return e.Code
}

func (e *Error) Detail() string {
	return e.Details
}

func ErrorCode(e error) string {
	if e == nil {
		return ""
	}
	err := e.(*Error)
	return err.Value()
}

func ErrorMessage(e error) string {
	if e == nil {
		return ""
	}
	err := e.(*Error)
	return err.Error()
}

func ErrorDetails(e error) string {
	if e == nil {
		return ""
	}
	err := e.(*Error)
	return err.Detail()
}

func ErrorNative(e error) error {
	if e == nil {
		return nil
	}
	err := e.(*Error)
	return err.Native
}
