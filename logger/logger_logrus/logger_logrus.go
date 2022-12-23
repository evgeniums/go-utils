package logger_logrus

import (
	"errors"
	"fmt"
	"os"

	"github.com/evgeniums/go-backend-helpers/config"
	"github.com/evgeniums/go-backend-helpers/logger"
	"github.com/evgeniums/go-backend-helpers/utils"
	"github.com/sirupsen/logrus"
)

type LogrusConfig struct {
	Destination string
	File        string
	LogLevel    string
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

func (l *LogrusLogger) Init(cfg config.Config, configPath string) error {

	// TODO load configuration
	l.Destination = cfg.GetString(config.Key(configPath, "destination"))
	l.File = cfg.GetString(config.Key(configPath, "file"))
	l.LogLevel = cfg.GetString(config.Key(configPath, "level"))

	l.Logrus = logrus.New()

	var err error
	if l.Destination == "file" {
		logFileName := l.File
		if logFileName == "" {
			logFileName = "log.log"
		}
		writer := &utils.FileWriteReopen{Path: logFileName}
		writer.File, err = os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			l.Logrus.Out = writer
			logrus.SetOutput(writer)
			fmt.Printf("Using log file %v\n", logFileName)
		} else {
			fmt.Println("Failed to log to file, using default console")
		}
	} else {
		l.Logrus.Out = os.Stdout
		logrus.SetOutput(os.Stdout)
	}
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
	return err
}

type WithLogrus struct {
	Log *LogrusLogger
}

func (w *WithLogrus) Logger() logger.Logger {
	return w.Log
}
