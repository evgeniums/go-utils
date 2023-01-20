package logger_logrus

import (
	"errors"
	"fmt"
	"os"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/config/object_config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"github.com/sirupsen/logrus"
)

type logrusConfig struct {
	DESTINATION string `default:"stdout" validate:"oneof=stdout file" vmessage:"logger destination must be one of: stdout | file"`
	FILE        string
	LEVEL       string `validate:"omitempty,oneof=panic fatal error warn info debug trace" vmessage:"invalid log level, must be one of: panic | fatal | error | warn | info | debug | trace"`
}

type LogrusLogger struct {
	logrusConfig
	logRus *logrus.Logger
}

func (l *LogrusLogger) Config() interface{} {
	return &l.logrusConfig
}

func New() *LogrusLogger {
	return &LogrusLogger{}
}

func (l *LogrusLogger) ErrorRaw(data ...interface{}) {
	l.logRus.Error(data)
}

func (l *LogrusLogger) Log(level logger.Level, message string, fields ...logger.Fields) {
	l.logRus.WithFields(logger.NewFields(fields...)).Log(logrus.Level(int(level)), message)
}

func (l *LogrusLogger) Debug(message string, fields ...logger.Fields) {
	l.logRus.WithFields(logger.NewFields(fields...)).Debug(message)
}

func (l *LogrusLogger) Trace(message string, fields ...logger.Fields) {
	l.logRus.WithFields(logger.NewFields(fields...)).Trace(message)
}

func (l *LogrusLogger) Error(message string, err error, fields ...logger.Fields) error {
	e := err
	if e == nil {
		if message != "" {
			e = errors.New(message)
		} else {
			e = errors.New("unknown error")
		}
	}
	f := logger.AppendFields(logger.Fields{"error": e}, fields...)
	if message != "" && err != nil {
		l.logRus.WithFields(f).Error(message)
	} else {
		l.logRus.WithFields(f).Error()
	}
	return e
}

func (l *LogrusLogger) Warn(message string, fields ...logger.Fields) {
	l.logRus.WithFields(logger.NewFields(fields...)).Warn(message)
}

func (l *LogrusLogger) Info(message string, fields ...logger.Fields) {
	l.logRus.WithFields(logger.NewFields(fields...)).Info(message)
}

func (l *LogrusLogger) Fatal(message string, err error, fields ...logger.Fields) error {
	e := err
	if e == nil {
		if message != "" {
			e = errors.New(message)
		} else {
			e = errors.New("unknown error")
		}
	}
	f := logger.AppendFields(logger.Fields{"error": e}, fields...)
	if message != "" && err != nil {
		l.logRus.WithFields(f).Log(logrus.FatalLevel, message)
	} else {
		l.logRus.WithFields(f).Log(logrus.FatalLevel)
	}
	return e
}

func (l *LogrusLogger) Init(cfg config.Config, vld validator.Validator, configPath ...string) error {

	// load configuration
	err := object_config.LoadValidate(cfg, vld, l, "logger", configPath...)
	if err != nil {
		return err
	}

	// setup logrus
	l.logRus = logrus.New()

	// setup output
	if l.DESTINATION == "file" {
		writer := &utils.FileWriteReopen{Path: l.FILE}
		writer.File, err = os.OpenFile(l.FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			l.logRus.Out = writer
			logrus.SetOutput(writer)
			fmt.Printf("Using log file %v\n", l.FILE)
		} else {
			fmt.Println("Failed to log to file, using default console")
		}
	} else {
		l.logRus.Out = os.Stdout
		logrus.SetOutput(os.Stdout)
	}

	// setup log level
	if l.LEVEL != "" {
		logLevel, err := logrus.ParseLevel(l.LEVEL)
		if err != nil {
			fmt.Printf("Invalid log level %v\n", err.Error())
		} else {
			fmt.Printf("Using log level %v\n", logLevel)
			l.logRus.SetLevel(logLevel)
			logrus.SetLevel(logLevel)
		}
	}

	// done
	return err
}

func (l *LogrusLogger) Native() interface{} {
	return l.logRus
}

func (l *LogrusLogger) ErrorNative(err error, fields ...logger.Fields) {
	f := logger.AppendFields(logger.Fields{"error": err}, fields...)
	l.logRus.WithFields(f).Error()
}

func (l *LogrusLogger) ErrorMessage(message string, fields ...logger.Fields) {
	err := errors.New(message)
	l.ErrorNative(err, fields...)
}
