package logger

import (
	"fmt"
	"os"

	"github.com/evgeniums/go-backend-helpers/utils"
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	Logrus *logrus.Logger
}

func (l *LogrusLogger) ErrorRaw(data ...interface{}) {
	l.Logrus.Error(data)
}

func (l *LogrusLogger) Debug(message string, fields ...Fields) {
	f := Fields{}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Debug(message)
}

func (l *LogrusLogger) Error(message string, err error, fields ...Fields) error {
	f := Fields{"error": err}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Error(message)

	return fmt.Errorf("%s: %s", message, err)
}

func (l *LogrusLogger) Warn(message string, fields ...Fields) {
	f := Fields{}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Warn(message)
}

func (l *LogrusLogger) Info(message string, fields ...Fields) {
	f := Fields{}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Info(message)
}

func (l *LogrusLogger) Fatal(message string, fields ...Fields) {
	f := Fields{}
	if len(fields) > 0 {
		f = utils.AppendMap(f, fields[0])
	}
	l.Logrus.WithFields(f).Fatal(message)
}

type LogrusConfig struct {
	Destination string
	File        string
	LogLevel    string
}

func (l *LogrusLogger) Init(config LogrusConfig) error {

	l.Logrus = logrus.New()

	var err error
	if config.Destination == "file" {
		logFileName := config.File
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
	if config.LogLevel != "" {
		logLevel, err := logrus.ParseLevel(config.LogLevel)
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

func (w *WithLogrus) Logger() Logger {
	return w.Log
}
