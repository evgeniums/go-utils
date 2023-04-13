package default_op_context

import (
	"errors"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/cache"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/generic_error"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/op_context"
	"github.com/evgeniums/go-backend-helpers/pkg/oplog"
	"github.com/evgeniums/go-backend-helpers/pkg/oplog/oplog_db"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type CallContextBase struct {
	method  string
	error_  error
	message string

	proxyLogger *logger.ProxyLogger
}

func (c *CallContextBase) SetLogger(logger.Logger) {}

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

func (c *CallContextBase) AddLoggerFields(fields logger.Fields) {
	for key, value := range fields {
		c.SetLoggerField(key, value)
	}
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
func (c *CallContextBase) SetErrorStr(err string) error {
	c.error_ = errors.New(err)
	return c.error_
}

func DefaultCallContextBuilder(methodName string, parentLogger logger.Logger, fields ...logger.Fields) op_context.CallContext {
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
	stack        []op_context.CallContext
	errorStack   []op_context.CallContext
	genericError generic_error.Error

	proxyLogger        *logger.ProxyLogger
	callContextBuilder op_context.CallContextBuilder
	cache              cache.Cache

	errorAsWarn bool

	oplogs       []oplog.Oplog
	oplogHandler op_context.OplogHandler

	origin        op_context.Origin
	writeCloseLog bool
}

func NewContext() *ContextBase {
	return &ContextBase{}
}

func (c *ContextBase) SetWriteCloseLog(enable bool) {
	c.writeCloseLog = enable
}

func (c *ContextBase) Init(app app_context.Context, log logger.Logger, db db.DB, fields ...logger.Fields) {

	c.writeCloseLog = true

	c.WithAppBase.Init(app)

	c.callContextBuilder = DefaultCallContextBuilder
	c.WithDBBase.Init(db)

	c.id = utils.GenerateID()

	staticLoggerFields := logger.AppendFieldsNew(logger.Fields{"context": c.id})
	c.proxyLogger = logger.NewProxy(log, logger.AppendFieldsNew(staticLoggerFields, fields...))
	c.WithLoggerBase.Init(c.proxyLogger)
	c.cache = app.Cache()

	c.stack = make([]op_context.CallContext, 0)

	c.oplogHandler = oplog_db.MakeOplogController

	c.Logger().Trace("open")
}

func (c *ContextBase) SetCallContextBuilder(builder op_context.CallContextBuilder) {
	c.callContextBuilder = builder
}

func (c *ContextBase) SetCache(cache cache.Cache) {
	c.cache = cache
}

func (c *ContextBase) SetOplogHandler(handler op_context.OplogHandler) {
	c.oplogHandler = handler
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

func stackPath(stack []op_context.CallContext) string {
	path := ""
	for _, method := range stack {
		if path != "" {
			path += ":"
		}
		path += method.Method()
	}
	return path
}

func (c *ContextBase) TraceInMethod(methodName string, fields ...logger.Fields) op_context.CallContext {

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
		c.errorStack = make([]op_context.CallContext, len(c.stack))
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
	if c.genericError == nil || err == nil || utils.OptionalArg(false, override...) {
		c.genericError = err
	}
}

func (c *ContextBase) GenericError() generic_error.Error {
	return c.genericError
}

func (c *ContextBase) DumpLog(successMessage ...string) {

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
		c.stack = []op_context.CallContext{}
		c.UnsetLoggerField("stack")
	} else {
		// log success
		if c.writeCloseLog {
			msg := utils.OptionalArg("success", successMessage...)
			if msg != "" {
				c.Logger().Info(msg)
			}
		}
	}

	c.ClearError()
}

func (c *ContextBase) Close(successMessage ...string) {

	// write oplog
	if c.oplogHandler != nil {
		oplog := c.oplogHandler(c)
		for _, o := range c.oplogs {
			o.InitObject()
			o.SetContext(c.ID())
			o.SetContextName(c.Name())
			if c.origin != nil {
				o.SetOriginApp(c.origin.App())
				o.SetOriginName(c.origin.Name())
				o.SetOriginSource(c.origin.Source())
				o.SetUser(c.origin.User())
				o.SetOriginClient(c.origin.SessionClient())
				o.SetUserType(c.origin.UserType())
			}
			oplog.Write(o)
		}
	}

	c.DumpLog(successMessage...)
}

func (c *ContextBase) AddLoggerFields(fields logger.Fields) {
	for key, value := range fields {
		c.SetLoggerField(key, value)
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

func (c *ContextBase) DbTransaction() db.Transaction {
	return c.dbTransaction
}

func (c *ContextBase) SetDbTransaction(tx db.Transaction) {
	c.dbTransaction = tx
}

func (c *ContextBase) ClearDbTransaction() {
	c.dbTransaction = nil
}

func (c *ContextBase) SetErrorAsWarn(enable bool) {
	c.errorAsWarn = enable
}

func (c *ContextBase) Reset() {
	c.stack = make([]op_context.CallContext, 0)
	c.errorStack = nil
	c.genericError = nil
	c.oplogs = make([]oplog.Oplog, 0)
}

func (c *ContextBase) ClearError() {
	c.errorStack = nil
	c.genericError = nil
}

func (c *ContextBase) Oplog(o oplog.Oplog) {
	if c.oplogs == nil {
		c.oplogs = make([]oplog.Oplog, 0)
	}
	c.oplogs = append(c.oplogs, o)
}

func (c *ContextBase) Origin() op_context.Origin {
	return c.origin
}

func (c *ContextBase) SetOrigin(o op_context.Origin) {
	c.origin = o
}

func (c *ContextBase) ExecDbTransaction(handler func() error) error {
	h := func(tx db.Transaction) error {

		c.SetDbTransaction(tx)
		defer c.ClearDbTransaction()

		return handler()
	}
	return c.Db().Transaction(h)
}

type OriginHolder struct {
	App           string
	Name          string
	Source        string
	SessionClient string
	User          string
	UserType      string
}

type Origin struct {
	OriginHolder
}

func NewOrigin(app app_context.Context) *Origin {
	o := &Origin{}
	o.SetApp(app.Application())
	o.SetName(app.AppInstance())
	o.SetSource(app.Hostname())
	return o
}

func (o *Origin) App() string {
	return o.OriginHolder.App
}

func (o *Origin) SetApp(val string) {
	o.OriginHolder.App = val
}

func (o *Origin) Name() string {
	return o.OriginHolder.Name
}

func (o *Origin) SetName(val string) {
	o.OriginHolder.Name = val
}

func (o *Origin) Source() string {
	return o.OriginHolder.Source
}

func (o *Origin) SetSource(val string) {
	o.OriginHolder.Source = val
}

func (o *Origin) SessionClient() string {
	return o.OriginHolder.SessionClient
}

func (o *Origin) SetSessionClient(val string) {
	o.OriginHolder.SessionClient = val
}

func (o *Origin) SetUser(val string) {
	o.OriginHolder.User = val
}

func (o *Origin) User() string {
	return o.OriginHolder.User
}

func (o *Origin) SetUserType(val string) {
	o.OriginHolder.UserType = val
}

func (o *Origin) UserType() string {
	return o.OriginHolder.UserType
}
