package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"github.com/tamago-cn/cmdline"
)

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 36
	gray    = 37
)

// LogFormatter 日志格式化
type LogFormatter struct {
	EnableTime      bool
	EnablePos       bool
	EnableColor     bool
	TimestampFormat string
	CallerLevel     int
}

func init() {
	cmdline.Regist("logger", "log", addLogger, "添加日志记录", `log <filename>`, []*cmdline.Param{
		&cmdline.Param{Name: "filename", Type: "string", Necessity: true, Desc: "日志文件名"},
	})
}

// Format renders a single log entry
func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	f.colored(b, entry, f.TimestampFormat)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *LogFormatter) colored(b *bytes.Buffer, entry *log.Entry, timestampFormat string) {
	var levelColor int
	switch entry.Level {
	case log.DebugLevel:
		levelColor = gray
	case log.WarnLevel:
		levelColor = yellow
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	// 封装层次较深
	//for i := 0; i < 20; i++ {
	//	_, file, line, ok := runtime.Caller(i)
	//	if !ok {
	//		file = "unknown"
	//		line = 0
	//	}
	//	fmt.Println(i, file, line)
	//}
	_, file, line, ok := runtime.Caller(f.CallerLevel)
	if !ok {
		file = "unknown"
		line = 0
	}
	file = path.Base(file)
	timePrefix := ""
	if f.EnableTime {
		timePrefix = fmt.Sprintf("%s ", entry.Time.Format(timestampFormat))
	}
	pos := ""
	if f.EnablePos {
		pos = fmt.Sprintf("[%s:%d] ", file, line)
	}
	levelText := strings.ToUpper(entry.Level.String())[0:4]
	if f.EnableColor {
		levelText = fmt.Sprintf("[\x1b[%dm%s\x1b[0m] ", levelColor, levelText)
	}

	fmt.Fprintf(b, "%s%s%s%-44s ", timePrefix, pos, levelText, entry.Message)
}

// addLogger 内置命令，增加日志记录
func addLogger(file string) {
	lfHook := lfshook.NewHook(
		file,
		&LogFormatter{
			EnableTime:      true,
			EnablePos:       true,
			EnableColor:     true,
			TimestampFormat: "2006-01-02 15:04:05",
			CallerLevel:     10,
		})
	log.AddHook(lfHook)
}

// setLogger 设置默认日志格式
func setLogger() {
	log.SetFormatter(&LogFormatter{
		EnableColor:     true,
		TimestampFormat: "",
		CallerLevel:     7,
	})
	log.SetOutput(os.Stdout)
}
