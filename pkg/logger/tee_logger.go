package logger

import "errors"

type TeeLogger struct {
	loggers []Logger
}

func NewTee(loggers ...Logger) *TeeLogger {
	l := &TeeLogger{}
	l.loggers = make([]Logger, 0)
	l.loggers = append(l.loggers, loggers...)
	return l
}

func (t *TeeLogger) ErrorRaw(data ...interface{}) {
	for _, logger := range t.loggers {
		logger.ErrorRaw(data)
	}
}

func (t *TeeLogger) Log(level Level, message string, fields ...Fields) {
	for _, logger := range t.loggers {
		logger.Log(level, message, fields...)
	}
}

func (t *TeeLogger) Debug(message string, fields ...Fields) {
	for _, logger := range t.loggers {
		logger.Debug(message, fields...)
	}
}

func (t *TeeLogger) Trace(message string, fields ...Fields) {
	for _, logger := range t.loggers {
		logger.Trace(message, fields...)
	}
}

func (t *TeeLogger) Error(message string, err error, fields ...Fields) error {
	for _, logger := range t.loggers {
		logger.Error(message, err, fields...)
	}
	if err == nil {
		if message != "" {
			return errors.New(message)
		} else {
			return errors.New("unknown error")
		}
	}
	return err
}

func (t *TeeLogger) ErrorNative(err error, fields ...Fields) {
	for _, logger := range t.loggers {
		logger.ErrorNative(err, fields...)
	}
}

func (t *TeeLogger) ErrorMessage(message string, fields ...Fields) {
	for _, logger := range t.loggers {
		logger.ErrorMessage(message, fields...)
	}
}

func (t *TeeLogger) Warn(message string, fields ...Fields) {
	for _, logger := range t.loggers {
		logger.Warn(message, fields...)
	}
}

func (t *TeeLogger) Info(message string, fields ...Fields) {
	for _, logger := range t.loggers {
		logger.Info(message, fields...)
	}
}

func (t *TeeLogger) Fatal(message string, err error, fields ...Fields) error {
	for _, logger := range t.loggers {
		logger.Fatal(message, err, fields...)
	}
	if err == nil {
		if message != "" {
			return errors.New(message)
		} else {
			return errors.New("unknown error")
		}
	}
	return err
}

func (t *TeeLogger) Native() interface{} {
	return t.loggers[0].Native()
}

func (t *TeeLogger) PushFatalStack(message string, err error, fields ...Fields) error {
	for _, logger := range t.loggers {
		logger.PushFatalStack(message, err, fields...)
	}
	if err == nil {
		if message != "" {
			return errors.New(message)
		} else {
			return errors.New("unknown error")
		}
	}
	return err
}

func (t *TeeLogger) CheckFatalStack(logger Logger, message ...string) bool {
	yes := false
	for _, logger := range t.loggers {
		yes_ := logger.CheckFatalStack(logger, message...)
		yes = yes || yes_
	}
	return yes
}

func (t *TeeLogger) DumpRequests() bool {
	for _, logger := range t.loggers {
		if logger.DumpRequests() {
			return true
		}
	}
	return false
}

func (t *TeeLogger) SetLevel(level Level) {
	for _, logger := range t.loggers {
		logger.SetLevel(level)
	}
}

func (t *TeeLogger) GetLevel() Level {
	for _, logger := range t.loggers {
		return logger.GetLevel()
	}
	return InfoLevel
}
