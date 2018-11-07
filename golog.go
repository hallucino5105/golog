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

type Output uint8
type Level uint8
type colorFunc func(...interface{}) string

const (
	unknownOutput Output = iota
	OStdout
	OStderr
)

const (
	unknownLevel Level = iota
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
var glcur *GoLog
var currentOutput Output = OStderr

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

func NewGoLog(output Output, option *GoLogOption) *GoLog {
	gl := new(GoLog)

	gl.Colorize = option.Colorize
	gl.MinLevel = option.MinLevel
	gl.DefaultLevel = LInfo
	gl.Header = nil
	gl.UserHeader = ""

	switch output {
	case OStdout:
		gl.out = os.Stdout
	case OStderr:
		gl.out = os.Stderr
	default:
		log.Panic("Output is unknown")
	}

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

	glstd = NewGoLog(OStdout, option)
	glerr = NewGoLog(OStderr, option)

	SetOutput(currentOutput)
}

func register(gl *GoLog) {
	gl.setDefaultHeader()
}

func SetOutput(output Output) {
	switch output {
	case OStdout:
		glcur = getStdLogger()
	case OStderr:
		glcur = getErrLogger()
	default:
		log.Panic("Output is unknown")
	}

	currentOutput = output
}

func GetCurrentOutput() Output {
	return currentOutput
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

func getCurrentLogger() *GoLog {
	if glcur == nil {
		SetOutput(currentOutput)
	}

	return glcur
}

func sprintf(text string, args []interface{}) string {
	return fmt.Sprintf(text, args...)
}

func Log(text string, args ...interface{}) {
	logger := getCurrentLogger()
	logger.write(getFormattedText(sprintf(text, args), logger, logger.DefaultLevel))
}

func Trace(text string, args ...interface{}) {
	logger := getCurrentLogger()
	logger.write(getFormattedText(sprintf(text, args), logger, LTrace))
}

func Debug(text string, args ...interface{}) {
	logger := getCurrentLogger()
	logger.write(getFormattedText(sprintf(text, args), logger, LDebug))
}

func Info(text string, args ...interface{}) {
	logger := getCurrentLogger()
	logger.write(getFormattedText(sprintf(text, args), logger, LInfo))
}

func Notice(text string, args ...interface{}) {
	logger := getCurrentLogger()
	logger.write(getFormattedText(sprintf(text, args), logger, LNotice))
}

func Warn(text string, args ...interface{}) {
	logger := getCurrentLogger()
	logger.write(getFormattedText(sprintf(text, args), logger, LWarning))
}

func Error(text string, args ...interface{}) {
	logger := getCurrentLogger()
	logger.write(getFormattedText(sprintf(text, args), logger, LError))
}

func Panic(text string, args ...interface{}) {
	logger := getCurrentLogger()
	logger.write(getFormattedText(sprintf(text, args), logger, LPanic))
	os.Exit(-1)
}
