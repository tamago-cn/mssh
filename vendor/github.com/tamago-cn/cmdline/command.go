package cmdline

import (
	"os"
	"os/exec"
	"runtime"

	log "github.com/sirupsen/logrus"
)

func init() {
	Regist("cmdline", "clear", Clear, "清屏", `clear`, []*Param{})
	Regist("cmdline", "exit", Exit, "退出", `exit`, []*Param{})
	Regist("cmdline", "vim", Vim, "打开vim编辑器", `vim <filename>`, []*Param{
		&Param{Name: "filename", Type: "string", Necessity: true, Desc: "文件名"},
	})
}

// Clear 内置命令，清屏
func Clear() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

// Vim 打开vim编辑器
func Vim(filename string) {
	if runtime.GOOS != "windows" {
		cmd := exec.Command("vim", filename)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			log.Errorf("open vim failed, error: %s", err.Error())
			return
		}
		log.Infof("vim edit success")
	} else {
		log.Warnf("windows does not support vim")
	}
}

// Exit 退出mssh命令行
func Exit() {
	os.Exit(0)
}
