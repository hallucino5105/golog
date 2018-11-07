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

	golog.Std.Log("aaa %d", 1)
	golog.Std.Trace("test")
	golog.Std.Debug("test")
	golog.Std.Info("test")
	golog.Std.Notice("test")
	golog.Std.Warn("test")
	golog.Std.Error("test")

	golog.Err.Log("bbb %d", 2)
	golog.Err.Trace("test")
	golog.Err.Debug("test")
	golog.Err.Info("test")
	golog.Err.Notice("test")
	golog.Err.Warn("test")
	golog.Err.Error("test")
}
