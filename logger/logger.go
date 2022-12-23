package logger

type Fields = map[string]interface{}

type Logger interface {
	Error(message string, err error, fields ...Fields) error
	Warn(message string, fields ...Fields)
	Debug(message string, fields ...Fields)
	Info(message string, fields ...Fields)
	Fatal(message string, err error, fields ...Fields) error

	ErrorRaw(...interface{})
}

type WithLogger interface {
	Logger() Logger
}

type WithLoggerBase struct {
	logger Logger
}

func (w *WithLoggerBase) Logger() Logger {
	return w.logger
}

func (w *WithLoggerBase) SetLogger(logger Logger) {
	w.logger = logger
}
