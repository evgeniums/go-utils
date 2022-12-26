package logger_logrus

import (
	"errors"
	"fmt"
	"os"

	"github.com/evgeniums/go-backend-helpers/pkg/config"
	"github.com/evgeniums/go-backend-helpers/pkg/logger"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/evgeniums/go-backend-helpers/pkg/validator"
	"github.com/sirupsen/logrus"
)

type LogrusConfig struct {
	Destination string `config:"destination" defaul:"stdout" validate:"oneof=stdout file" vmessage:"logger destination must be one of: stdout | file"`
	File        string `config:"file" validate:"required" vmessage:"path of logger file must be set"`
	LogLevel    string `config:"log_level" validate:"omitempty,oneof=panic fatal error warn info debug trace" vmessage:"invalid log level, must be one of: panic | fatal | error | warn | info | debug | trace"`
}

type LogrusLogger struct {
	LogrusConfig
	Logrus *logrus.Logger
}

func New() *LogrusLogger {
	return &LogrusLogger{}
}

func (l *LogrusLogger) ErrorRaw(data ...interface{}) {
	l.Logrus.Error(data)
}

func (l *LogrusLogger) Debug(message string, fields ...logger.Fields) {
	f := logger.Fields{}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Debug(message)
}

func (l *LogrusLogger) Error(message string, err error, fields ...logger.Fields) error {
	f := logger.Fields{"error": err}
	if err == nil {
		f = logger.Fields{}
	}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Error(message)

	if err == nil {
		return errors.New(message)
	}

	return fmt.Errorf("%s: %s", message, err)
}

func (l *LogrusLogger) Warn(message string, fields ...logger.Fields) {
	f := logger.Fields{}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Warn(message)
}

func (l *LogrusLogger) Info(message string, fields ...logger.Fields) {
	f := logger.Fields{}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Info(message)
}

func (l *LogrusLogger) Fatal(message string, err error, fields ...logger.Fields) error {
	f := logger.Fields{"error": err}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Fatal(message)
	return fmt.Errorf("%s: %s", message, err)
}

func (l *LogrusLogger) Init(cfg config.Config, vld validator.Validator, configPath ...string) error {

	// load configuration
	err := config.InitObjectNoLog(cfg, vld, l, "logger", configPath...)
	if err != nil {
		return err
	}

	// setup logrus
	l.Logrus = logrus.New()

	// setup output
	if l.Destination == "file" {
		writer := &utils.FileWriteReopen{Path: l.File}
		writer.File, err = os.OpenFile(l.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			l.Logrus.Out = writer
			logrus.SetOutput(writer)
			fmt.Printf("Using log file %v\n", l.File)
		} else {
			fmt.Println("Failed to log to file, using default console")
		}
	} else {
		l.Logrus.Out = os.Stdout
		logrus.SetOutput(os.Stdout)
	}

	// setup log level
	if l.LogLevel != "" {
		logLevel, err := logrus.ParseLevel(l.LogLevel)
		if err != nil {
			fmt.Printf("Invalid log level %v\n", err.Error())
		} else {
			fmt.Printf("Using log level %v\n", logLevel)
			l.Logrus.SetLevel(logLevel)
			logrus.SetLevel(logLevel)
		}
	}

	// done
	return err
}
