package logger

import "github.com/sirupsen/logrus"

type Fields map[string]interface{}

type LoggerLevel string

const (
	DebugLevel LoggerLevel = "debug"
	InfoLevel  LoggerLevel = "info"
	WarnLevel  LoggerLevel = "warn"
	ErrorLevel LoggerLevel = "error"
	FatalLevel LoggerLevel = "fatal"
	PanicLevel LoggerLevel = "panic"
)

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	WithFields(fields Fields) Logger
}

type logger struct {
	*logrus.Logger
}

func New(level LoggerLevel) Logger {
	logrusLevel, err := logrus.ParseLevel(string(level))
	if err != nil {
		logrusLevel = logrus.InfoLevel
	}

	lgr := logrus.New()
	lgr.SetLevel(logrusLevel)

	lgr.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	return logger{lgr}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{
		Logger: l.Logger.WithFields(logrus.Fields(fields)).Logger,
	}
}

var GlobalLogger Logger

func init() {
	GlobalLogger = New(InfoLevel)
}
