package generic_error

// Generic error that can be forwarded from place of arising to place of user reporting.
type Error interface {
	error
	Code() string
	Message() string
	Details() string
	Original() error
}

type ErrorBase struct {
	code     string
	message  string
	details  string
	original error
}

const UnknownError string = "unknown_error"

// Create new error from code and message.
func New(code string, message ...string) *ErrorBase {
	e := &ErrorBase{code: code}
	if len(message) > 0 {
		e.message = message[0]
	}
	return e
}

// Create new error from code and message taken from other "native error".
func NewFromErr(code string, err error) *ErrorBase {
	return New(code, err.Error())
}

// Create new error from code, message and some other "original error" with keeping native error.
func NewFromOriginal(code string, message string, original error) *ErrorBase {
	e := &ErrorBase{code: code, message: message, original: original}
	return e
}

// Create new error from message.
func NewFromMessage(message string) *ErrorBase {
	e := &ErrorBase{code: UnknownError, message: message}
	return e
}

// Convert error to string for error interface.
func (e *ErrorBase) Error() string {
	return e.message
}

// Get error code.
func (e *ErrorBase) Code() string {
	return e.code
}

// Get error details.
func (e *ErrorBase) Details() string {
	return e.details
}

// Get original error.
func (e *ErrorBase) Original() error {
	return e.original
}

// Extract code from the error. If error is not of Error type then code is unknown_error.
func Code(e error) string {
	if e == nil {
		return ""
	}
	err, ok := e.(Error)
	if !ok {
		return UnknownError
	}
	return err.Code()
}

// Extract message from the error. If error is not of Error type then error as string is used.
func Message(e error) string {
	if e == nil {
		return ""
	}
	err, ok := e.(Error)
	if !ok {
		return e.Error()
	}
	return err.Error()
}

// Extract details from the error.
func Details(e error) string {
	if e == nil {
		return ""
	}
	err, ok := e.(Error)
	if !ok {
		return ""
	}
	return err.Details()
}

// Extract original error from the error. If error is not of Error type then the argument is returned as is.
func Original(e error) error {
	if e == nil {
		return nil
	}
	err, ok := e.(Error)
	if !ok {
		return err
	}
	return err.Original()
}
