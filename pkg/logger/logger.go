package logger

import (
	"errors"

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

	PushFatalStack(message string, err error, fields ...Fields) error
	CheckFatalStack(Logger)
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

func AppendFields(f Fields, fields ...Fields) {
	if len(fields) > 0 {
		utils.AppendMap(f, fields[0])
	}
}

func AppendFieldsNew(f Fields, fields ...Fields) Fields {
	newFields := utils.CopyMap(f)
	if len(fields) > 0 {
		utils.AppendMap(newFields, fields[0])
	}
	return newFields
}

func NewFields(fields ...Fields) Fields {
	if len(fields) > 0 {
		return utils.CopyMap(fields[0])
	}
	return Fields{}
}

type fatalError struct {
	messageStack []string
	deepestError error
	fields       Fields
}

type LoggerBase struct {
	fatalError fatalError
}

func (l *LoggerBase) Init() {
	l.fatalError.messageStack = make([]string, 0)
	l.fatalError.fields = NewFields()
	l.fatalError.deepestError = nil
}

func (l *LoggerBase) CheckFatalStack(logger Logger) {
	if l.fatalError.deepestError != nil {
		errMsg := ""
		for i := len(l.fatalError.messageStack) - 1; i >= 0; i-- {
			msg := l.fatalError.messageStack[i]
			errMsg += msg
			if i != 0 {
				errMsg += ": "
			}
		}
		logger.Fatal(errMsg, l.fatalError.deepestError, l.fatalError.fields)
		l.Init()
	}
}

func (l *LoggerBase) PushFatalStack(message string, err error, fields ...Fields) error {
	if l.fatalError.messageStack == nil {
		l.Init()
	}

	e := err
	if e == nil {
		if message == "" {
			e = errors.New("unknown error")
		} else {
			e = errors.New(message)
		}
	}

	if l.fatalError.deepestError == nil {
		l.fatalError.deepestError = e
	}

	if message != "" {
		l.fatalError.messageStack = append(l.fatalError.messageStack, message)
	} else {
		l.fatalError.messageStack = append(l.fatalError.messageStack, e.Error())
	}

	AppendFields(l.fatalError.fields, fields...)

	return e
}
