package logger

import (
	"log"
	"os"
	"path"
	"time"

	"bitbucket.org/linkernetworks/aurora/src/config"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

// the singleton logger
var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
}

// Setup Logger in packge. Enable Logger after import
func Setup(c config.LoggerConfig) {
	if Logger == nil {
		Logger = logrus.New()
	}
	configure(Logger, c)
}

// New a logger in scope
func New(c config.LoggerConfig) *logrus.Logger {
	var logger = logrus.New()
	configure(logger, c)
	return logger
}

func configure(logger *logrus.Logger, c config.LoggerConfig) {
	logger.Formatter = new(prefixed.TextFormatter)

	// preparing log dir
	dir := c.Dir

	if dir == "" {
		dir = "log"
		logger.Infof("Log dir is not specified. Using default log dir: %s", dir)
	}

	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}

	linkName := c.LinkName
	if linkName == "" {
		linkName = "access_log"
	}
	logger.Infof("Start writing log to %s", path.Join(dir, linkName))

	filePattern := linkName + c.SuffixPattern
	if filePattern == "" || c.SuffixPattern == "" {
		filePattern = "access_log.%Y%m%d"
	}

	var maxAge = 24 * time.Hour
	if d, err := time.ParseDuration(c.MaxAge); err == nil {
		maxAge = d
	}
	logger.Infof("Max age of logs: %s", maxAge.String())

	writer, err := rotatelogs.New(
		path.Join(dir, filePattern),
		rotatelogs.WithLinkName(path.Join(dir, linkName)),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
	)
	if err != nil {
		log.Panic(err)
	}

	logger.Hooks.Add(
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
		logger.SetLevel(logrus.ErrorLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
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
