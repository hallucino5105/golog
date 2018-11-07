package golog

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"
)

type GoLog struct {
	MinLevel     Level
	DefaultLevel Level
	Colorize     bool
	Header       *template.Template
	UserHeader   string

	mu  sync.Mutex
	out io.Writer
}

type GoLogOption struct {
	Colorize bool
	MinLevel Level
}

type HeaderDefaultParam struct {
	Level  string
	Date   string
	Caller string
}

type StdOutput struct {
	logger *GoLog
}

type ErrOutput struct {
	logger *GoLog
}

type Level uint8
type colorFunc func(...interface{}) string

const (
	unknown Level = iota
	LTrace
	LDebug
	LInfo
	LNotice
	LWarning
	LError
	LPanic
)

var glstd *GoLog
var glerr *GoLog

var Std *StdOutput
var Err *ErrOutput

func (level Level) Color() colorFunc {
	switch level {
	case LTrace:
		return color.New(color.FgWhite).SprintFunc()
	case LDebug:
		return color.New(color.FgBlue).SprintFunc()
	case LInfo:
		return color.New(color.FgGreen).SprintFunc()
	case LNotice:
		return color.New(color.FgMagenta).SprintFunc()
	case LWarning:
		return color.New(color.FgYellow).SprintFunc()
	case LError:
		return color.New(color.FgRed).SprintFunc()
	case LPanic:
		return color.New(color.FgHiWhite, color.BgRed).SprintFunc()
	}

	return color.New(color.FgWhite).SprintFunc()
}

func (level Level) String() string {
	switch level {
	case LTrace:
		return " trace"
	case LDebug:
		return " debug"
	case LInfo:
		return "  info"
	case LNotice:
		return "notice"
	case LWarning:
		return "  warn"
	case LError:
		return " error"
	case LPanic:
		return " panic"
	}

	return "unknown"
}

func NewGoLog(out io.Writer, option *GoLogOption) *GoLog {
	gl := new(GoLog)

	gl.Colorize = option.Colorize
	gl.MinLevel = option.MinLevel
	gl.DefaultLevel = LInfo
	gl.Header = nil
	gl.UserHeader = ""
	gl.out = out

	register(gl)

	return gl
}

func SetupLogger(option *GoLogOption) {
	if option == nil {
		option = &GoLogOption{
			Colorize: true,
			MinLevel: LDebug,
		}
	}

	glstd = NewGoLog(os.Stdout, option)
	glerr = NewGoLog(os.Stderr, option)

	Std = &StdOutput{logger: getStdLogger()}
	Err = &ErrOutput{logger: getErrLogger()}
}

func register(gl *GoLog) {
	gl.setDefaultHeader()
}

func (gl *GoLog) setDefaultHeader() {
	tmplStr := "[{{.Level}}] {{.Date}} ({{.Caller}}): "
	tmpl, err := template.New("GoLogHeaderTemplate").Parse(tmplStr)
	if err != nil {
		panic(err)
	}

	gl.Header = tmpl
}

func (gl *GoLog) SetUserHeader(header string) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	gl.UserHeader = header
}

func (gl *GoLog) SetMinLevel(level Level) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	gl.MinLevel = level
}

func (gl *GoLog) SetDefaultLevel(level Level) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	gl.DefaultLevel = level
}

func (gl *GoLog) SetColorize(colorize bool) {
	gl.mu.Lock()
	defer gl.mu.Unlock()

	gl.Colorize = colorize
}

func (gl *GoLog) write(text string, level Level) {
	if level >= gl.MinLevel {
		gl.out.Write([]byte(text + "\n"))
	}
}

func getDate() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func getCaller(logger *GoLog) string {
	var caller string = "unknown"

	_, sourceFileName, sourceFileLineNum, ok := runtime.Caller(4)
	if ok {
		if logger.Colorize {
			caller = color.CyanString(
				fmt.Sprintf("%s:%d", filepath.Base(sourceFileName), sourceFileLineNum))
		} else {
			caller = fmt.Sprintf("%s:%d", filepath.Base(sourceFileName), sourceFileLineNum)
		}
	}

	return caller
}

func getHeader(logger *GoLog, level Level) string {
	var header string
	if logger.UserHeader != "" {
		header = logger.UserHeader
	} else {
		var levelStr string = level.String()
		if logger.Colorize {
			levelStr = level.Color()(levelStr)
		}

		hp := HeaderDefaultParam{
			Level:  levelStr,
			Date:   getDate(),
			Caller: getCaller(logger),
		}

		var buf bytes.Buffer
		logger.Header.Execute(&buf, hp)

		header = buf.String()
	}

	return header
}

func getFormattedText(text string, logger *GoLog, level Level) (string, Level) {
	header := getHeader(logger, level)
	return header + text, level
}

func sprintf(text string, args []interface{}) string {
	return fmt.Sprintf(text, args...)
}

func getStdLogger() *GoLog {
	if glstd == nil {
		log.Panic("The logger object is not initialized. Please call SetupLogger().")
	}

	return glstd
}

func getErrLogger() *GoLog {
	if glstd == nil {
		log.Panic("The logger object is not initialized. Please call SetupLogger().")
	}

	return glerr
}

func (o *StdOutput) Log(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, o.logger.DefaultLevel))
}

func (o *StdOutput) Trace(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LTrace))
}

func (o *StdOutput) Debug(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LDebug))
}

func (o *StdOutput) Info(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LInfo))
}

func (o *StdOutput) Notice(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LNotice))
}

func (o *StdOutput) Warn(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LWarning))
}

func (o *StdOutput) Error(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LError))
}

func (o *StdOutput) Panic(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LPanic))
	os.Exit(-1)
}

func (o *ErrOutput) Log(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, o.logger.DefaultLevel))
}

func (o *ErrOutput) Trace(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LTrace))
}

func (o *ErrOutput) Debug(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LDebug))
}

func (o *ErrOutput) Info(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LInfo))
}

func (o *ErrOutput) Notice(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LNotice))
}

func (o *ErrOutput) Warn(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LWarning))
}

func (o *ErrOutput) Error(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LError))
}

func (o *ErrOutput) Panic(text string, args ...interface{}) {
	o.logger.write(getFormattedText(sprintf(text, args), o.logger, LPanic))
	os.Exit(-1)
}
