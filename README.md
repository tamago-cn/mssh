# mssh
基于golang (golang.org/x/crypto/ssh) 的批量远程工具，针对多台服务器远程批量执行相同指令，附带文件上传下载的功能，提升服务部署效率

## 构建
go install

## 内置命令
    log: 日志记录
    put: 上传文件
    get: 下载文件
    done: 等待一组并行命令结束，在执行并行命令，如connect之后，使用此命令以保证后续步骤的正常执行
    connect: 连接远程服务器，可并行
    release: 释放指定host的远程连接
    check: 检查已建立连接
    clear: 清屏
    run: 执行mssh脚本(可将指令集整理成脚本文件，使用此命令执行)
    exit: 退出(或使用Ctrl+C)

## 初始化加载
    mssh在初始化时会默认加载执行目录下的.msshrc文件并解释执行，一些初始化操作比如预先连接的服务器可在此文件中记录，以达到便捷配置的效果
    比如:
    log test.log
    connect <username1> <passwd1> <host1>
    conenct <username2> <passwd1> <host2>
    done
    check

    可以配置记录日志到test.log
    并且启动连接服务器host1,host2
    等待所有连接执行完成
    进行一次检测，输出连接情况


## 使用
    * 命令行执行：mssh 开启mssh命令行
    [mssh ~]#

    * 查看命令帮助
    [mssh ~]# help <command(可选，默认显示命令列表)>

    * 开启一个日志
    [mssh ~]# log <filename>

    * 连接远程服务器
    [mssh ~]# connect <username> <passwd> <host> <port(可选，默认22)>

    * 等待并行命令结束
    [mssh ~]# done

    * 释放远程连接
    [mssh ~]# release <host>

    * 检查已建立连接
    [mssh ~]# check

    * 执行远程命令，支持除mssh内置命令外的所有远程机器支持的命令，同时执行多条指令需用";"隔开
    [mssh ~]# ls
    [mssh ~]# cd /tmp;pwd


    * 上传文件:
    [mssh ~]# put <localPath> <remoteDir(可选，默认用户主目录，如:/root/)>
    针对get命令下载的文件，localPath使用'@/<文件名>'的形式可将download下对应host下的的文件上传

    * 下载文件:
    [mssh ~]# get <远程文件>
    (文件下载至./download/<host>/下)

    * 执行mssh脚本
    [mssh ~]# run <mssh脚本>

    * 清屏
    [mssh ~]# clear

    * 退出
    [mssh ~]# exit




