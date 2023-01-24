package logger

import (
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
)

type Fields = map[string]interface{}
type Level int

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
	// TraceLevel level. Designates finer-grained informational events than the Debug.
	TraceLevel
)

type Logger interface {
	Log(level Level, message string, fields ...Fields)

	Error(message string, err error, fields ...Fields) error
	ErrorNative(err error, fields ...Fields)
	ErrorMessage(message string, fields ...Fields)

	Warn(message string, fields ...Fields)
	Debug(message string, fields ...Fields)
	Info(message string, fields ...Fields)
	Trace(message string, fields ...Fields)
	Fatal(message string, err error, fields ...Fields) error

	ErrorRaw(...interface{})

	Native() interface{}
}

type WithLogger interface {
	Logger() Logger
	SetLogger(logger Logger)
}

type WithLoggerBase struct {
	logger Logger
}

func (w *WithLoggerBase) Logger() Logger {
	return w.logger
}

func (w *WithLoggerBase) Init(logger Logger) {
	w.logger = logger
}

func (w *WithLoggerBase) SetLogger(logger Logger) {
	w.logger = logger
}

func AppendFields(f Fields, fields ...Fields) Fields {
	newFields := utils.CopyMap(f)
	if len(fields) > 0 {
		newFields = utils.AppendMap(newFields, fields[0])
	}
	return newFields
}

func NewFields(fields ...Fields) Fields {
	if len(fields) > 0 {
		return utils.CopyMap(fields[0])
	}
	return Fields{}
}
