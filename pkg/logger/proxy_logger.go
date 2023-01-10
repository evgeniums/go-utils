package logger

import (
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type ProxyLogger struct {
	logger Logger
	fields Fields
}

func NewProxy(logger Logger, fields ...Fields) *ProxyLogger {
	return &ProxyLogger{logger, utils.OptionalArg(Fields{}, fields...)}
}

func (p *ProxyLogger) SetNextLogger(logger Logger) {
	p.logger = logger
}

func (p *ProxyLogger) SetField(name string, value interface{}) {
	p.fields[name] = value
}

func (p *ProxyLogger) StaticFields() Fields {
	return p.fields
}

func (p *ProxyLogger) UnsetField(name string) {
	delete(p.fields, name)
}

func (p *ProxyLogger) ErrorRaw(data ...interface{}) {
	p.logger.ErrorRaw(data)
}

func (p *ProxyLogger) Log(level Level, message string, fields ...Fields) {
	p.logger.Log(level, message, AppendFields(p.fields, fields...))
}

func (p *ProxyLogger) Debug(message string, fields ...Fields) {
	p.logger.Debug(message, AppendFields(p.fields, fields...))
}

func (p *ProxyLogger) Trace(message string, fields ...Fields) {
	p.logger.Trace(message, AppendFields(p.fields, fields...))
}

func (p *ProxyLogger) Error(message string, err error, fields ...Fields) error {
	return p.logger.Error(message, err, AppendFields(p.fields, fields...))
}

func (p *ProxyLogger) Warn(message string, fields ...Fields) {
	p.logger.Warn(message, AppendFields(p.fields, fields...))
}

func (p *ProxyLogger) Info(message string, fields ...Fields) {
	p.logger.Info(message, AppendFields(p.fields, fields...))
}

func (p *ProxyLogger) Fatal(message string, err error, fields ...Fields) error {
	return p.logger.Fatal(message, err, AppendFields(p.fields, fields...))
}

func (p *ProxyLogger) Native() interface{} {
	return p.logger.Native()
}
