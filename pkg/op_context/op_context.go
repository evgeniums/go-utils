package op_context

import (
	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/common"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/oplog"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type CallContext interface {
	Method() string
	Error() error
	Message() string

	SetError(err error) error
	SetMessage(msg string)

	SetLoggerField(name string, value interface{})
	UnsetLoggerField(name string)
	LoggerFields() logger.Fields

	Logger() logger.Logger
}

type CallContextBase struct {
	method  string
	error_  error
	message string

	proxyLogger *logger.ProxyLogger
}

func (c *CallContextBase) Method() string {
	return c.method
}
func (c *CallContextBase) Error() error {
	return c.error_
}
func (c *CallContextBase) Err() *error {
	return &c.error_
}
func (c *CallContextBase) Message() string {
	return c.message
}

func (c *CallContextBase) SetLoggerField(name string, value interface{}) {
	c.proxyLogger.SetStaticField(name, value)
}

func (c *CallContextBase) LoggerFields() logger.Fields {
	return c.proxyLogger.StaticFields()
}

func (c *CallContextBase) UnsetLoggerField(name string) {
	c.proxyLogger.UnsetStaticField(name)
}

func (c *CallContextBase) Logger() logger.Logger {
	return c.proxyLogger
}
func (c *CallContextBase) SetError(err error) error {
	c.error_ = err
	return c.error_
}
func (c *CallContextBase) SetMessage(msg string) {
	c.message = msg
}

type Context interface {
	app_context.WithApp
	common.WithName
	logger.WithLogger
	db.WithDB

	MainDB() db.DB
	MainLogger() logger.Logger

	DbTransaction() db.DBHandlers

	Cache() cache.Cache

	ErrorManager() generic_error.ErrorManager
	SetErrorManager(manager generic_error.ErrorManager)

	MakeGenericError(code string) generic_error.Error

	ID() string

	TraceInMethod(methodName string, fields ...logger.Fields) CallContext
	TraceOutMethod()

	SetGenericError(err generic_error.Error, override ...bool)
	GenericError() generic_error.Error
	SetGenericErrorCode(code string, override ...bool)

	Tr(phrase string) string

	SetLoggerField(name string, value interface{})
	LoggerFields() logger.Fields
	UnsetLoggerField(name string)

	SetErrorAsWarn(enable bool)

	Oplog(o oplog.Oplog)

	Reset()
	Close(successMessage ...string)
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
	app_context.WithAppBase
	logger.WithLoggerBase
	db.WithDBBase

	dbTransaction db.Transaction
	errorManager  generic_error.ErrorManager

	id           string
	name         string
	stack        []CallContext
	errorStack   []CallContext
	genericError generic_error.Error

	proxyLogger        *logger.ProxyLogger
	callContextBuilder CallContextBuilder
	cache              cache.Cache

	errorAsWarn bool

	oplogs []oplog.Oplog
}

func (c *ContextBase) Init(app app_context.Context, log logger.Logger, db db.DB, fields ...logger.Fields) {

	c.WithAppBase.Init(app)

	c.callContextBuilder = DefaultCallContextBuilder
	c.WithDBBase.Init(db)

	c.id = utils.GenerateID()

	staticLoggerFields := logger.AppendFieldsNew(logger.Fields{"context": c.id})
	c.proxyLogger = logger.NewProxy(log, logger.AppendFieldsNew(staticLoggerFields, fields...))
	c.WithLoggerBase.Init(c.proxyLogger)
	c.cache = app.Cache()

	c.stack = make([]CallContext, 0)
	c.Logger().Trace("open")
}

func (c *ContextBase) SetCallContextBuilder(builder CallContextBuilder) {
	c.callContextBuilder = builder
}

func (c *ContextBase) SetCache(cache cache.Cache) {
	c.cache = cache
}

func (c *ContextBase) SetErrorManager(manager generic_error.ErrorManager) {
	c.errorManager = manager
}

func (c *ContextBase) ID() string {
	return c.id
}

func (c *ContextBase) MainDB() db.DB {
	return c.WithDBBase.Db()
}

func (c *ContextBase) Name() string {
	return c.name
}

func (c *ContextBase) MainLogger() logger.Logger {
	return c.proxyLogger.NextLogger()
}

func (c *ContextBase) SetName(name string) {
	c.name = name
	c.SetLoggerField("op", c.name)
	c.Logger().Trace("name")
}

func (c *ContextBase) Tr(phrase string) string {
	return phrase
}

func stackPath(stack []CallContext) string {
	path := ""
	for _, method := range stack {
		if path != "" {
			path += ":"
		}
		path += method.Method()
	}
	return path
}

func (c *ContextBase) TraceInMethod(methodName string, fields ...logger.Fields) CallContext {

	ctx := c.callContextBuilder(methodName, c.proxyLogger, fields...)

	c.stack = append(c.stack, ctx)
	c.SetLoggerField("stack", stackPath(c.stack))

	c.Logger().Trace("callin")

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

	c.Logger().Trace("callout")

	if len(c.stack) == 0 {
		return
	}

	if c.stack[len(c.stack)-1].Error() != nil && c.errorStack == nil {
		c.errorStack = make([]CallContext, len(c.stack))
		copy(c.errorStack, c.stack)
	}

	c.stack = c.stack[:len(c.stack)-1]
	if len(c.stack) == 0 {
		c.UnsetLoggerField("stack")
	} else {
		c.SetLoggerField("stack", stackPath(c.stack))
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

func (c *ContextBase) Close(successMessage ...string) {

	// TODO write oplog

	if c.errorStack != nil {
		// log error
		var msg string
		var err error
		for _, item := range c.errorStack {
			// collect messages
			if item.Message() != "" {
				if msg != "" {
					msg = utils.ConcatStrings(msg, ":")
				}
				msg = utils.ConcatStrings(msg, item.Method(), "(", item.Message(), ")")
			}
			if item.Error() != nil {
				// override with deepest error
				err = item.Error()
			}
		}
		c.stack = c.errorStack
		c.SetLoggerField("stack", stackPath(c.stack))
		if !c.errorAsWarn {
			c.Logger().Error(msg, err)
		} else {
			c.Logger().Warn(msg, logger.Fields{"error": err})
		}
		c.stack = []CallContext{}
		c.UnsetLoggerField("stack")
	} else {
		// log success
		msg := utils.OptionalArg("success", successMessage...)
		if msg != "" {
			c.Logger().Info(msg)
		}
	}
}

func (c *ContextBase) SetLoggerField(name string, value interface{}) {
	c.proxyLogger.SetStaticField(name, value)
}

func (c *ContextBase) LoggerFields() logger.Fields {
	return c.proxyLogger.StaticFields()
}

func (c *ContextBase) UnsetLoggerField(name string) {
	c.proxyLogger.UnsetStaticField(name)
}

func (c *ContextBase) ErrorManager() generic_error.ErrorManager {
	return c.errorManager
}

func (c *ContextBase) MakeGenericError(code string) generic_error.Error {
	return c.errorManager.MakeGenericError(code, c.Tr)
}

func (c *ContextBase) SetGenericErrorCode(code string, override ...bool) {
	c.SetGenericError(c.MakeGenericError(code), override...)
}

func (c *ContextBase) Cache() cache.Cache {
	return c.cache
}

func (c *ContextBase) DbTransaction() db.DBHandlers {
	return c.dbTransaction
}

func (c *ContextBase) SetErrorAsWarn(enable bool) {
	c.errorAsWarn = enable
}

func (c *ContextBase) Reset() {
	c.stack = make([]CallContext, 0)
	c.errorStack = nil
	c.genericError = nil
	c.oplogs = make([]oplog.Oplog, 0)
}

func (c *ContextBase) Oplog(o oplog.Oplog) {
	if c.oplogs == nil {
		c.oplogs = make([]oplog.Oplog, 0)
	}
	c.oplogs = append(c.oplogs, o)
}

func DB(c Context) db.DBHandlers {
	if c.DbTransaction() != nil {
		return c.DbTransaction()
	}
	return c.Db()
}
