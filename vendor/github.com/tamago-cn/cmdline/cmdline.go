package cmdline

import (
	"fmt"
	"io/ioutil"

	"github.com/chzyer/readline"
)

// funcMap 命令方法映射
var funcMap map[string]*FuncInfo

func init() {
	funcMap = make(map[string]*FuncInfo)
	Regist("cmdline", "help", Help, "显示帮助信息", `help <command>`, []*Param{
		&Param{Name: "command", Type: "string", Necessity: false, Desc: "指令名"},
	})
}

// Param 参数信息
type Param struct {
	Name      string
	Type      string
	Necessity bool
	Desc      string
}

// FuncInfo 方法信息
type FuncInfo struct {
	Group  string
	Name   string
	Fn     interface{}
	Help   string
	Usage  string
	Params []*Param
}

// Regist 注册命令行方法
func Regist(group string, name string, fn interface{}, help string, usage string, params []*Param) {
	if _, ok := funcMap[name]; ok {
		return
	}
	f := FuncInfo{
		Group:  group,
		Name:   name,
		Fn:     fn,
		Help:   help,
		Usage:  usage,
		Params: params,
	}
	funcMap[name] = &f
}

// Help 显示帮助
func Help(name string) {
	if name == "" {
		// 按group分组
		groups := map[string]map[string]*FuncInfo{}
		for k, f := range funcMap {
			if _, ok := groups[f.Group]; ok {
				groups[f.Group][k] = f
			} else {
				groups[f.Group] = make(map[string]*FuncInfo)
				groups[f.Group][k] = f
			}
		}
		// 显示简要帮助信息
		fmt.Println("所有指令:")
		for gn, group := range groups {
			fmt.Printf("  [%s]\n", gn)
			for k, f := range group {
				fmt.Printf("    %s: %s\n", k, f.Help)
			}
		}
		return
	}
	if f, ok := funcMap[name]; ok {
		fmt.Printf("指令: %s\n", f.Name)
		fmt.Printf("功能: %s\n", f.Help)
		fmt.Printf("用法: %s\n", f.Usage)
		fmt.Printf("参数:\n")
		for _, p := range f.Params {
			n := "必填"
			if p.Necessity == false {
				n = "可选"
			}
			fmt.Printf("    <%s>  (%s), %s, %s\n", p.Name, p.Type, n, p.Desc)
		}
	} else {
		fmt.Printf("    方法 %s 未注册\n", name)
	}
}

// listFiles 列出某目录下的文件，以用作tab补全
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

// listFuncs 列出某目录下的文件，以用作tab补全
func listFuncs() func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		for _, f := range funcMap {
			names = append(names, f.Name)
		}
		return names
	}
}

// GetCompleter 获取自动补全方法
func GetCompleter() *readline.PrefixCompleter {
	top := []readline.PrefixCompleterInterface{}
	cmds := []string{"python", "sh", "bash"}
	for _, cmd := range cmds {
		top = append(top, readline.PcItem(cmd, readline.PcItemDynamic(listFiles("./"))))
	}
	for k, f := range funcMap {
		sub := []readline.PrefixCompleterInterface{}
		for _, p := range f.Params {
			sub = append(sub, readline.PcItem(fmt.Sprintf("<%s>", p.Name)))
		}
		sub = append(sub, readline.PcItemDynamic(listFiles("./")))
		sub = append(sub, readline.PcItemDynamic(listFuncs()))
		top = append(top, readline.PcItem(k, sub...))
	}
	completer := readline.NewPrefixCompleter(top...)
	return completer
}

// GetFunc 获取执行方法
func GetFunc(name string) (interface{}, error) {
	if f, ok := funcMap[name]; ok {
		return f.Fn, nil
	}
	return nil, fmt.Errorf("func '%s' not regist", name)
}
