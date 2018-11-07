package golog_test

import (
	"testing"

	"github.com/miyaizu/golog"
)

func TestGoLog(t *testing.T) {
	golog.SetupLogger(&golog.GoLogOption{
		Colorize: true,
		MinLevel: golog.LTrace,
	})

	golog.SetOutput(golog.OStdout)
	golog.Log("aaa %d", 1)
	golog.Trace("test")
	golog.Debug("test")
	golog.Info("test")
	golog.Notice("test")
	golog.Warn("test")
	golog.Error("test")

	golog.SetOutput(golog.OStderr)
	golog.Log("aaa %d", 1)
	golog.Trace("test")
	golog.Debug("test")
	golog.Info("test")
	golog.Notice("test")
	golog.Warn("test")
	golog.Error("test")
}
