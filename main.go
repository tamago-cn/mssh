package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/tamago-cn/cmdline"

	_ "github.com/tamago-cn/mssh/pkg/cmd"
)

var (
	// VERSION 版本信息
	VERSION string
	// BUILD 构建时间
	BUILD string
	// COMMITSHA1 git commit ID
	COMMITSHA1 string
)

//命令行选项
var (
	h = flag.Bool("h", false, "Show this help")
	v = flag.Bool("v", false, "Show version")
)

func init() {
	setLogger()
}

func main() {
	// 加载初始化配置
	initFile := ".msshrc"
	cmdline.Setup("mssh", "/tmp/mssh.tmp")
	cmdline.Run(initFile)

	log.Infoln("main start")
	if len(os.Args) > 1 {
		// 若带参数，则解释文件
		for _, file := range os.Args[1:] {
			cmdline.Run(file)
		}
	} else {
		cmdline.Interpret(os.Stdin)
	}
}
