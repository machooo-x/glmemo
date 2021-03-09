package syslog

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Clog Clog
var Clog *LogAttributes

const (
	// LogError enable error level
	LogError int = 0x01
	// LogWarn enable warnning level
	LogWarn int = 0x02
	// LogTrace enable trace level
	LogTrace int = 0x04
	// LogInfo enable info level
	LogInfo int = 0x08
	// LogQueueSize for cache queue size
	LogQueueSize int = 5000
)

// LogAttributes Log Attributes
type LogAttributes struct {
	ServiceName string
	Flag        int
	LogIO       *lumberjack.Logger
}

func init() {
	logpacker := &lumberjack.Logger{
		Filename:   "maximo.log",
		MaxSize:    10, //MB
		MaxBackups: 10,
		MaxAge:     28, //days
	}

	Clog = &LogAttributes{
		Flag:  0,
		LogIO: logpacker,
	}
}

// SetLogServiceName 初始化调用
func (la *LogAttributes) SetLogServiceName(svrname string) {
	if svrname == "" {
		svrname = "defaultSvr"
	}
	la.ServiceName = svrname
}

// Infoln Info with ln.
// stdout bool 是否输出到控制台
func (la *LogAttributes) Infoln(stdout bool, a ...interface{}) {
	logLoc := getLogFileInfo()
	if la.Flag&LogInfo != 0 {
		color.New(color.FgHiRed).Fprint(la.LogIO, "["+time.Now().Format("2006-01-02 15:04:05.000")+"] "+logLoc+" ")
		color.New(color.FgGreen).Fprintln(la.LogIO, a...)
	}
	if stdout {
		color.New(color.FgHiRed).Fprint(os.Stdout, "["+time.Now().Format("2006-01-02 15:04:05.000")+"] "+logLoc+" ")
		color.New(color.FgGreen).Fprintln(os.Stdout, a...)
	}
}

// Info Info without ln.
// stdout bool 是否输出到控制台
func (la *LogAttributes) Info(stdout bool, a ...interface{}) {
	if la.Flag&LogInfo != 0 {
		color.New(color.FgGreen).Fprint(la.LogIO, a...)
	}
	if stdout {
		color.New(color.FgGreen).Fprint(os.Stdout, a...)
	}
}

// Errorln Error with ln.
// stdout bool 是否输出到控制台
func (la *LogAttributes) Errorln(stdout bool, a ...interface{}) {
	logLoc := getLogFileInfo()
	if la.Flag&LogError != 0 {
		color.New(color.FgHiRed).Fprint(la.LogIO, "["+time.Now().Format("2006-01-02 15:04:05.000")+"] "+logLoc+" ")
		color.New(color.FgHiRed).Fprintln(la.LogIO, a...)
	}
	if stdout {
		color.New(color.FgHiRed).Fprint(os.Stdout, "["+time.Now().Format("2006-01-02 15:04:05.000")+"] "+logLoc+" ")
		color.New(color.FgHiRed).Fprintln(os.Stdout, a...)
	}
}

// Error Error without ln.
// stdout bool 是否输出到控制台
func (la *LogAttributes) Error(stdout bool, a ...interface{}) {
	if la.Flag&LogError != 0 {
		color.New(color.FgHiRed).Fprint(la.LogIO, a...)
	}
	if stdout {
		color.New(color.FgHiRed).Fprint(os.Stdout, a...)
	}
}

// Warnln Warn with ln.
// stdout bool 是否输出到控制台
func (la *LogAttributes) Warnln(stdout bool, a ...interface{}) {
	logLoc := getLogFileInfo()
	if la.Flag&LogWarn != 0 {
		color.New(color.FgHiRed).Fprint(la.LogIO, "["+time.Now().Format("2006-01-02 15:04:05.000")+"] "+logLoc+" ")
		color.New(color.FgYellow).Fprintln(la.LogIO, a...)
	}
	if stdout {
		color.New(color.FgHiRed).Fprint(os.Stdout, "["+time.Now().Format("2006-01-02 15:04:05.000")+"] "+logLoc+" ")
		color.New(color.FgYellow).Fprintln(os.Stdout, a...)
	}
}

// Warn Warn without ln.
// stdout bool 是否输出到控制台
func (la *LogAttributes) Warn(stdout bool, a ...interface{}) {
	if la.Flag&LogWarn != 0 {
		color.New(color.FgYellow).Fprintln(la.LogIO, a...)
	}
	if stdout {
		color.New(color.FgYellow).Fprintln(os.Stdout, a...)
	}
}

// Traceln Trace with ln.
// stdout bool 是否输出到控制台
func (la *LogAttributes) Traceln(stdout bool, a ...interface{}) {
	logLoc := getLogFileInfo()
	if la.Flag&LogTrace != 0 {
		color.New(color.FgHiRed).Fprint(la.LogIO, "["+time.Now().Format("2006-01-02 15:04:05.000")+"] "+logLoc+" ")
		color.New(color.FgCyan).Fprintln(la.LogIO, a...)
	}
	if stdout {
		color.New(color.FgHiRed).Fprint(os.Stdout, "["+time.Now().Format("2006-01-02 15:04:05.000")+"] "+logLoc+" ")
		color.New(color.FgCyan).Fprintln(os.Stdout, a...)
	}
}

// Trace Trace without ln.
// stdout bool 是否输出到控制台
func (la *LogAttributes) Trace(stdout bool, a ...interface{}) {
	if la.Flag&LogTrace != 0 {
		color.New(color.FgCyan).Fprintln(la.LogIO, a...)
	}
	if stdout {
		color.New(color.FgCyan).Fprintln(os.Stdout, a...)
	}
}

// 获取日志打印文件和行数
func getLogFileInfo() string {
	f := "???"
	n := "0"

	_, file, line, ok := runtime.Caller(2)
	if ok {
		indx := strings.LastIndex(file, "/")
		if indx > 0 {
			f = file[indx+1:]
			n = fmt.Sprint(line)
		}
	}
	return f + ":" + fmt.Sprint(n)
}
