package generic_error

import (
	"net/http"

	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type TranslationHandler = func(string) string

type ErrorManager interface {
	ErrorDescription(code string, tr ...TranslationHandler) string
	ErrorProtocolCode(code string) int

	MakeGenericError(code string, tr ...TranslationHandler) Error

	AddErrorDescriptions(m map[string]string)
	AddErrorProtocolCodes(m map[string]int)

	SetDefaultErrorProtocolCode(code int)
	DefaultErrorProtocolCode() int
}

type ErrorDefinitions interface {
	AttachToErrorManager(manager ErrorManager)
}

type ErrorManagerBase struct {
	descriptions        map[string]string
	protocolCodes       map[string]int
	defaultProtocolCode int
}

func (e *ErrorManagerBase) Init(defaultProtocolCode int) {
	e.defaultProtocolCode = defaultProtocolCode
	e.AddErrorDescriptions(CommonErrorDescriptions)
}

func (e *ErrorManagerBase) DefaultErrorProtocolCode() int {
	return e.defaultProtocolCode
}

func (e *ErrorManagerBase) SetDefaultErrorProtocolCode(code int) {
	e.defaultProtocolCode = code
}

func (e *ErrorManagerBase) AddErrorDescriptions(m map[string]string) {
	utils.AppendMap(e.descriptions, m)
}

func (e *ErrorManagerBase) AddErrorProtocolCodes(m map[string]int) {
	utils.AppendMap(e.protocolCodes, m)
}

func (e *ErrorManagerBase) ErrorDescription(code string, tr ...TranslationHandler) string {
	description, ok := e.descriptions[code]
	if !ok {
		description = code
	}
	if len(tr) > 0 {
		description = tr[0](description)
	}
	return description
}

func (e *ErrorManagerBase) ErrorProtocolCode(code string) int {
	protocolCode, ok := e.protocolCodes[code]
	if !ok {
		return e.defaultProtocolCode
	}
	return protocolCode
}

func (e *ErrorManagerBase) MakeGenericError(code string, tr ...TranslationHandler) Error {
	err := New(code, e.ErrorDescription(code, tr...))
	return err
}

type ErrorManagerBaseHttp struct {
	ErrorManagerBase
}

func (e *ErrorManagerBaseHttp) Init() {
	e.ErrorManagerBase.Init(http.StatusBadRequest)
	e.AddErrorProtocolCodes(CommonErrorHttpCodes)
}

type ErrorsExtender interface {
	AddToErrorManager(manager ErrorManager)
}

type ErrorsExtenderBase struct {
	errorDescriptions  map[string]string
	errorProtocolCodes map[string]int
}

func (e *ErrorsExtenderBase) Init(errorDescriptions map[string]string, errorProtocolCodes ...map[string]int) {
	e.errorDescriptions = errorDescriptions
	if len(errorProtocolCodes) > 0 {
		e.errorProtocolCodes = errorProtocolCodes[0]
	}
}

func (e *ErrorsExtenderBase) AddToErrorManager(manager ErrorManager) {
	manager.AddErrorDescriptions(e.errorDescriptions)
	manager.AddErrorProtocolCodes(e.errorProtocolCodes)
}
