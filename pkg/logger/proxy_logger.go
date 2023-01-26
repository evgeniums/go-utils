package logger

import (
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ProxyLogger struct {
	logger       Logger
	staticFields Fields
}

func NewProxy(logger Logger, fields ...Fields) *ProxyLogger {
	return &ProxyLogger{logger, utils.OptionalArg(Fields{}, fields...)}
}

func (p *ProxyLogger) NextLogger() Logger {
	return p.logger
}

func (p *ProxyLogger) SetNextLogger(logger Logger) {
	p.logger = logger
}

func (p *ProxyLogger) SetStaticField(name string, value interface{}) {
	p.staticFields[name] = value
}

func (p *ProxyLogger) StaticFields() Fields {
	return p.staticFields
}

func (p *ProxyLogger) UnsetStaticField(name string) {
	delete(p.staticFields, name)
}

func (p *ProxyLogger) ErrorRaw(data ...interface{}) {
	p.logger.ErrorRaw(data)
}

func (p *ProxyLogger) Log(level Level, message string, fields ...Fields) {
	p.logger.Log(level, message, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) Debug(message string, fields ...Fields) {
	p.logger.Debug(message, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) Trace(message string, fields ...Fields) {
	p.logger.Trace(message, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) Error(message string, err error, fields ...Fields) error {
	return p.logger.Error(message, err, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) ErrorNative(err error, fields ...Fields) {
	p.logger.ErrorNative(err, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) ErrorMessage(message string, fields ...Fields) {
	p.logger.ErrorMessage(message, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) Warn(message string, fields ...Fields) {
	p.logger.Warn(message, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) Info(message string, fields ...Fields) {
	p.logger.Info(message, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) Fatal(message string, err error, fields ...Fields) error {
	return p.logger.Fatal(message, err, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) Native() interface{} {
	return p.logger.Native()
}

func (p *ProxyLogger) PushFatalStack(message string, err error, fields ...Fields) error {
	return p.logger.PushFatalStack(message, err, AppendFieldsNew(p.staticFields, fields...))
}

func (p *ProxyLogger) CheckFatalStack(logger Logger, message ...string) bool {
	return p.logger.CheckFatalStack(logger, message...)
}
