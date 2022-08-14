package logger

type Fields = map[string]interface{}

type Logger interface {
	Error(message string, err error, fields ...Fields)
	Warn(message string, fields ...Fields)
	Debug(message string, fields ...Fields)
	Info(message string, fields ...Fields)
	Fatal(message string, fields ...Fields)
	ErrorRaw(...interface{})
}

type WithLogger interface {
	Logger() Logger
}

type WithLoggerBase struct {
	LoggerInterface Logger
}

func (w *WithLoggerBase) Logger() Logger {
	return w.LoggerInterface
}
