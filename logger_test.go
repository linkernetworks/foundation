package logger

import "testing"

func TestInfo(t *testing.T) {
	Info("test")
}

func TestInfoln(t *testing.T) {
	Infoln("test", "line2")
}

func TestDebug(t *testing.T) {
	Debug("test", "line2")
}
