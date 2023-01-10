package op_context

import (
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type CallContext interface {
	Method() string
	Error() error
	Message() string
	Fields() logger.Fields

	Logger() logger.Logger
}

type CallContextBase struct {
	method  string
	error_  error
	message string

	proxyLogger *logger.ProxyLogger
}

func (t *CallContextBase) Method() string {
	return t.method
}
func (t *CallContextBase) Error() error {
	return t.error_
}
func (t *CallContextBase) Message() string {
	return t.message
}
func (t *CallContextBase) Fields() logger.Fields {
	return t.proxyLogger.StaticFields()
}
func (t *CallContextBase) Logger() logger.Logger {
	return t.proxyLogger
}

type Context interface {
	common.WithName
	logger.WithLogger
	db.WithDB

	ID() string

	TraceInMethod(methodName string, ctxBuilder ...CallContextBuilder) CallContext
	TraceOutMethod()

	SetGenericError(err generic_error.Error, override ...bool)
	GenericError() generic_error.Error

	Tr(phrase string) string

	Close()
}

type WithCtx interface {
	Ctx() Context
}

type CallContextBuilder = func(methodName string, parentLogger logger.Logger, fields ...logger.Fields) CallContext

func DefaultCallContextBuilder(methodName string, parentLogger logger.Logger, fields ...logger.Fields) CallContext {
	ctx := &CallContextBase{method: methodName, proxyLogger: logger.NewProxy(parentLogger, fields...)}
	return ctx
}

type ContextBase struct {
	logger.WithLoggerBase
	db.WithDBBase

	id           string
	name         string
	stack        []CallContext
	errorStack   []CallContext
	genericError generic_error.Error

	proxyLogger        *logger.ProxyLogger
	callContextBuilder CallContextBuilder
}

func (c *ContextBase) Init(log logger.Logger, db db.DB, fields ...logger.Fields) {

	c.callContextBuilder = DefaultCallContextBuilder
	c.WithDBBase.Init(db)

	c.id = utils.GenerateID()
	c.stack = make([]CallContext, 0)

	staticLoggerFields := logger.AppendFields(logger.Fields{"op_context": c.id})
	c.proxyLogger = logger.NewProxy(log, logger.AppendFields(staticLoggerFields, fields...))
	c.WithLoggerBase.Init(c.proxyLogger)

	c.stack[len(c.stack)-1].Logger().Trace("open op context")
}

func (c *ContextBase) SetCallContextBuilder(builder CallContextBuilder) {
	c.callContextBuilder = builder
}

func (c *ContextBase) ID() string {
	return c.id
}

func (c *ContextBase) Name() string {
	return c.name
}

func (c *ContextBase) SetName(name string) {
	c.name = name
	c.proxyLogger.SetField("op", c.name)
	c.Logger().Trace("name op context")
}

func (c *ContextBase) Tr(phrase string) string {
	return phrase
}

func stackPath(stack []CallContext) string {
	path := ""
	for _, method := range stack {
		if path == "" {
			path += ":"
		}
		path += method.Method()
	}
	return path
}

func (c *ContextBase) TraceInMethod(methodName string, fields ...logger.Fields) CallContext {

	ctx := c.callContextBuilder(methodName, c.proxyLogger, fields...)

	c.stack = append(c.stack, ctx)
	c.proxyLogger.SetField("stack", stackPath(c.stack))

	c.Logger().Trace("begin")

	return ctx
}

func (c *ContextBase) Logger() logger.Logger {
	if len(c.stack) == 0 {
		return c.proxyLogger
	}
	lastLogger := c.stack[len(c.stack)-1].Logger()
	if lastLogger != nil {
		return lastLogger
	}
	return c.proxyLogger
}

func (c *ContextBase) TraceOutMethod() {

	c.Logger().Trace("end")

	if len(c.stack) == 0 {
		return
	}

	if c.stack[len(c.stack)-1].Error() != nil && c.errorStack == nil {
		c.errorStack = make([]CallContext, len(c.stack))
		copy(c.errorStack, c.stack)
	}

	c.stack = c.stack[:len(c.stack)-1]
	if len(c.stack) == 0 {
		c.proxyLogger.UnsetField("stack")
	} else {
		c.proxyLogger.SetField("stack", stackPath(c.stack))
	}
}

func (c *ContextBase) SetGenericError(err generic_error.Error, override ...bool) {
	if c.genericError == nil || utils.OptionalArg(false, override...) {
		c.genericError = err
	}
}

func (c *ContextBase) GenericError() generic_error.Error {
	return c.genericError
}

func (c *ContextBase) Close() {

	// log errors
	if c.errorStack != nil {
		var msg string
		var err error
		for _, item := range c.errorStack {
			// collect messages
			if item.Message() != "" {
				if msg != "" {
					msg += ":"
				}
				msg += fmt.Sprintf("%s( %s )", item.Method(), item.Message())
			}
			if item.Error() != nil {
				// override with deepest error
				err = item.Error()
			}
		}
		c.stack = c.errorStack
		c.proxyLogger.SetField("stack", stackPath(c.stack))
		c.Logger().Error(msg, err)
		c.stack = []CallContext{}
		c.proxyLogger.UnsetField("stack")
	}

	c.Logger().Trace("close op context")
}
