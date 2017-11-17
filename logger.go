package logger

import (
	"log"
	"os"
	"path"
	"time"

	"bitbucket.org/linkernetworks/cv-tracker/server/config"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// the singleton logger
var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.Formatter = new(prefixed.TextFormatter)
}

func New(c config.LoggerConfig) *logrus.Logger {
	// preparing log dir
	dir := c.Dir

	if dir == "" {
		dir = "log"
		log.Println("Log dir is not specified. Using default log dir:", dir)
	}

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}

	filePattern := c.FilePattern
	if filePattern == "" {
		filePattern = "access_log.%Y%m%d"
	}

	linkName := c.LinkName
	if linkName == "" {
		linkName = "access_log"
	}

	log.Println("Start writing log to", path.Join(dir, linkName))
	writer, err := rotatelogs.New(
		path.Join(dir, filePattern),
		rotatelogs.WithLinkName(path.Join(dir, linkName)),
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

	switch c.Level {
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

	return Logger
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
