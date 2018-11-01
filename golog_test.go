package golog_test

import (
	"testing"

	"github.com/miyaizu/golog"
)

func TestGoLog(t *testing.T) {
	golog.SetupLogger(&golog.GoLogOption{
		Colorize: true,
		MinLevel: golog.LDebug,
	})

	golog.SLog("aaa %d", 1)
	golog.STrace("test")
	golog.SDebug("test")
	golog.SInfo("test")
	golog.SWarn("test")
	golog.SError("test")

	golog.ELog("bbb %d", 2)
	golog.ETrace("test")
	golog.EDebug("test")
	golog.EInfo("test")
	golog.EWarn("test")
	golog.EError("test")
}
