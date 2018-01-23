package logger

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLogger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	logger.Info("test")
	logger.Infoln("test", "line2")
	logger.Debug("test", "line2")
}
