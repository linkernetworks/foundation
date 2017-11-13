package logger

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// the singleton logger
var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.Formatter = new(prefixed.TextFormatter)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infoln(args ...interface{}) {
	Logger.Infoln(args...)
}

func Infof(msg string, args ...interface{}) {
	Logger.Infof(msg, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnln(args ...interface{}) {
	Logger.Warnln(args...)
}

func Warnf(msg string, args ...interface{}) {
	Logger.Warnf(msg, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorln(args ...interface{}) {
	Logger.Errorln(args...)
}

func Errorf(msg string, args ...interface{}) {
	Logger.Errorf(msg, args...)
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugln(args ...interface{}) {
	Logger.Debugln(args...)
}

func Debugf(msg string, args ...interface{}) {
	Logger.Debugf(msg, args...)
}

func Panic(args ...interface{}) {
	Logger.Panic(args...)
}

func Panicln(args ...interface{}) {
	Logger.Panicln(args...)
}

func Panicf(msg string, args ...interface{}) {
	Logger.Panicf(msg, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalln(args ...interface{}) {
	Logger.Fatalln(args...)
}

func Fatalf(msg string, args ...interface{}) {
	Logger.Fatalf(msg, args...)
}
