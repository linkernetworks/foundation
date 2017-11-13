package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"bitbucket.org/linkernetworks/cv-tracker/server/config"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

// the singleton logger
var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
}

func Setup(cf *config.Config) {
	Logger.Formatter = &logrus.TextFormatter{
		DisableSorting:  true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logDir := cf.App.LogDir
	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}

	// TODO hook to console and log file
	// TODO symlink
	writer, err := rotatelogs.New(
		fmt.Sprintf("%s%s.%s", logDir, "/cv-tracker.log", "%Y%m%d"),
		rotatelogs.WithLinkName(logDir),
		rotatelogs.WithMaxAge(24*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
	)
	if err != nil {
		log.Panic(err)
	}

	Logger.Hooks.Add(
		lfshook.NewHook(
			lfshook.WriterMap{
				logrus.DebugLevel: writer,
				logrus.InfoLevel:  writer,
				logrus.WarnLevel:  writer,
				logrus.ErrorLevel: writer,
				logrus.FatalLevel: writer,
			},
		),
	)

	switch cf.App.LogLevel {
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	default:
		Logger.SetLevel(logrus.InfoLevel)
	}
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
