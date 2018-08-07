package cmdline

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/chzyer/readline"
	"github.com/cosiner/argv"
	log "github.com/sirupsen/logrus"
)

func init() {
	conf = &Conf{
		Prompt:      "cmdline",
		HistoryFile: "/tmp/cmdline.tmp",
	}
	defaultFunc = func(line string) {
		fmt.Println("exec with defaultFunc:", line)
	}
	Regist("cmdline", "run", Run, "执行脚本文件", `run <filename>`, []*Param{
		&Param{Name: "filename", Type: "string", Necessity: true, Desc: "脚本文件名"},
	})
}

var (
	conf        *Conf
	defaultFunc func(string)
)

// Setup 解释器设置
func Setup(prompt string, historyFile string, fn func(string)) {
	conf.Prompt = prompt
	conf.HistoryFile = historyFile
	defaultFunc = fn
}

// Conf 解释器配置
type Conf struct {
	Prompt      string `ini:"prompt"`
	HistoryFile string `ini:"history_file"`
}

// call 内置命令通用调用方法
func call(fn interface{}, params ...interface{}) {
	ft := reflect.TypeOf(fn)
	extra := ft.NumIn() - len(params)
	if extra > 0 {
		// 函数需求参数多余提供参数的情况
		for i := 0; i < extra; i++ {
			params = append(params, "")
		}
	} else if extra < 0 {
		// 函数需求参数少于提供的参数的情况
		log.Warn("params count not match")
		return
	}
	f := reflect.ValueOf(fn)
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		//fmt.Println(param)
		in[k] = reflect.ValueOf(param)
	}
	f.Call(in)
}

// filterInput 过滤特定输入
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// Run 内置命令，执行mssh脚本
func Run(script string) {
	fp, err := os.Open(script)
	if err != nil {
		log.Errorf("run script %s error: %s", script, err.Error())
		return
	}
	defer fp.Close()
	Interpret(fp)
}

// Interpret 解释输入流
func Interpret(in io.ReadCloser) {
	readline.Stdin = in
	r, err := readline.NewEx(&readline.Config{
		Prompt:          fmt.Sprintf("\033[33m[%s ~ ]#\033[0m ", conf.Prompt),
		HistoryFile:     conf.HistoryFile,
		AutoComplete:    GetCompleter(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer r.Close()

	//log.SetOutput(r.Stderr())
	for {
		line, err := r.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		//parseLine(line)
		if len(line) == 0 {
			// 输入为空继续
			continue
		}
		if strings.HasPrefix(line, "#") {
			// 将"#"定为注释
			continue
		}
		args, err := argv.Argv([]rune(line), map[string]string{}, argv.Run)
		if err != nil {
			// 解析错误打日志继续
			log.Error(line, err)
			continue
		}
		if fn, err := GetFunc(args[0][0]); err == nil {
			params := []interface{}{}
			if len(args[0]) > 1 {
				for _, param := range args[0][1:] {
					params = append(params, param)
				}
			}
			call(fn, params...)
		} else if strings.HasPrefix(line, "run") {
			// 对run特殊处理，放在funcMap中会出现循环调用
			if len(args[0]) == 1 {
				continue
			}
			for _, param := range args[0][1:] {
				Run(param)
			}
		} else {
			defaultFunc(line)
		}
	}
}
