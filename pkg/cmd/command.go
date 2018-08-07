package cmd

import (
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tamago-cn/cmdline"
	"github.com/tamago-cn/mssh/pkg/scp"
	"golang.org/x/crypto/ssh"
)

func init() {
	cliMutex = &sync.Mutex{}
	cmdline.Regist("inner", "done", Done, "等待批量任务完成", `done`, []*cmdline.Param{})

	cmdline.Regist("file", "put", Put, "批量上传文件", `put <filePath> <remoteDir>`, []*cmdline.Param{
		&cmdline.Param{Name: "filePath", Type: "string", Necessity: true, Desc: "本地文件路径"},
		&cmdline.Param{Name: "remoteDir", Type: "string", Necessity: false, Desc: "远程目录, 默认用户home目录, 如: /root/"},
	})
	cmdline.Regist("file", "get", Get, "批量下载文件, 本操作会将文件下载到执行目录下的download目录下服务器地址对应目录中", `get <remotePath>`, []*cmdline.Param{
		&cmdline.Param{Name: "remotePath", Type: "string", Necessity: true, Desc: "远程文件路径"},
	})

	cmdline.Regist("conn", "check", Check, "检查连接状态", `check`, []*cmdline.Param{})
	cmdline.Regist("conn", "connect", Connect, "连接远程主机", `connect <username> <password> <host> <port> <timeout>`, []*cmdline.Param{
		&cmdline.Param{Name: "username", Type: "string", Necessity: true, Desc: "用户名"},
		&cmdline.Param{Name: "password", Type: "string", Necessity: true, Desc: "密码"},
		&cmdline.Param{Name: "host", Type: "string", Necessity: true, Desc: "服务器地址"},
		&cmdline.Param{Name: "port", Type: "int", Necessity: false, Desc: "sshd服务端口, 默认 22"},
		&cmdline.Param{Name: "timeout", Type: "int", Necessity: false, Desc: "连接超时时间(单位 s), 默认 5 "},
	})
	cmdline.Regist("conn", "release", Release, "释放远程连接", `release <host>`, []*cmdline.Param{
		&cmdline.Param{Name: "host", Type: "string", Necessity: true, Desc: "服务器地址"},
	})
}

// Client 一个ssh客户端的相关参数集合
type Client struct {
	Cli      *ssh.Client
	HomePath string
}

var (
	cliMutex *sync.Mutex
	cliMap   = map[string]*Client{}
	wg       sync.WaitGroup
)

// 启动一个并行任务
func launch(f func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		f()
	}()
}

// Done 内置命令，等待并行任务完成
func Done() {
	wg.Wait()
	log.Info("multi command done")
}

// Connect 内置命令，连接服务器
func Connect(user, password, host, port, timeout string) {
	launch(func() {
		// 此处为了防止建立连接耗时过长的情况，另开一个线程去处理
		if _, ok := cliMap[host]; ok {
			// 已连接的服务器不再重复连接
			log.Infof("[%s] connected", host)
			return
		}
		if port == "" {
			port = "22"
		}
		if timeout == "" {
			timeout = "5"
		}
		t, err := strconv.Atoi(timeout)
		if err != nil {
			log.Errorf("[%s] convert timeout to int error: %s", host, err.Error())
			return
		}
		var (
			auth         []ssh.AuthMethod
			addr         string
			clientConfig *ssh.ClientConfig
			client       *ssh.Client
			//session      *ssh.Session
			//err error
		)
		// get auth method
		auth = make([]ssh.AuthMethod, 0)
		auth = append(auth, ssh.Password(password))

		clientConfig = &ssh.ClientConfig{
			User:    user,
			Auth:    auth,
			Timeout: time.Duration(t) * time.Second,
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}

		addr = fmt.Sprintf("%s:%s", host, port)

		if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
			log.Errorf("[%s] ssh Dial error: %s", host, err)
			return
		}
		session, err := client.NewSession()
		if err != nil {
			log.Errorf("[%s] get session error: %s", host, err.Error())
			return
		}
		homePath, err := session.Output("pwd")
		if err != nil {
			log.Errorf("[%s] get home path error: %s", host, err.Error())
			return
		}
		//session.Stdout = os.Stdout
		cli := &Client{
			Cli:      client,
			HomePath: strings.TrimSpace(string(homePath)),
		}

		cliMutex.Lock()
		defer cliMutex.Unlock()
		cliMap[host] = cli
		log.Infof("[%s] connect success", host)
	})
}

// Release 内置命令，释放连接
func Release(host string) {
	if client, ok := cliMap[host]; ok {
		err := client.Cli.Close()
		if err != nil {
			log.Warnf("[%s] close error: %s", host, err)
		}
		delete(cliMap, host)
		log.Warnf("[%s] released", host)
		return
	}
	log.Warnf("[%s] has not connected yet", host)
}

// Put 内置命令，批量上传文件
func Put(file string, dstDir string) {
	//fmt.Println("func put:", file, dstDir)
	for host, client := range cliMap {
		wg.Add(1)
		go func(host string, client *Client, file, dstDir string) {
			defer wg.Done()
			toDir := dstDir
			if toDir == "" {
				toDir = client.HomePath
			}
			remoteFileName := path.Base(file)
			remotePath := path.Join(toDir, remoteFileName)
			session, err := client.Cli.NewSession()
			if err != nil {
				log.Errorf("[%s] get session error: %s", host, err.Error())
				return
			}
			// 针对下载的备份做特殊处理
			file = strings.Replace(file, "@", fmt.Sprintf("download/%s", host), -1)
			//err = scp.CopyPath(file, remotePath, session)
			scp.CopyToRemote(session, file, remotePath)
			if err != nil {
				log.Errorf("[%s] scp file %s error: %s", host, file, err)
				return
			}
			log.Infof("[%s] put file [%s] to [%s] success", host, file, remotePath)
		}(host, client, file, dstDir)
	}
	Done()
}

// Get 内置命令，批量下载文件
func Get(file string) {
	for host, client := range cliMap {
		wg.Add(1)
		go func(host string, client *Client, file string) {
			defer wg.Done()
			session, err := client.Cli.NewSession()
			if err != nil {
				log.Errorf("[%s] get session error: %s", host, err)
				return
			}
			localDir := path.Join(".", "download", host)
			localFileName := path.Base(file)
			os.MkdirAll(localDir, 0666)
			localPath := path.Join(localDir, localFileName)
			err = scp.CopyFromRemote(session, file, localPath)
			if err != nil {
				log.Errorf("[%s] copy from remote error: %s", host, err)
				return
			}
			log.Infof("[%s] get file [%s] to [%s] success", host, file, localPath)
		}(host, client, file)
	}
	Done()
}

// Check 内置命令，检测已建立连接
func Check() {
	for host, client := range cliMap {
		_ = client
		log.Infof("[%s] connecting", host)
	}
}

// Remote 内置命令，批量远程执行
func Remote(cmd string) {
	for host, client := range cliMap {
		fmt.Printf("\033[36m>>>>>>>>>>>>>>> %s [%s] <<<<<<<<<<<<<<<\033[0m\n", host, cmd)
		session, err := client.Cli.NewSession()
		//session.Stdin = os.Stdin
		session.Stdout = os.Stdout
		//session.Stderr = os.Stderr
		if err != nil {
			log.Errorf("[%s] get session error: %s\n", host, err.Error())
			continue
		}
		err = session.Run(cmd)
		if err != nil {
			// 此次执行有错误
			log.Errorf("[%s] remote command [%s] failed: %s\n", host, cmd, err.Error())
			continue
		}
		log.Infof("[%s] remote command [%s] success\n", host, cmd)
		session.Close()
	}
}
