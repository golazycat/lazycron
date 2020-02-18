package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/golazycat/lazycron/common"

	"github.com/golazycat/lazycron/common/conf"

	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/master"
)

// 初始化基本runtime环境
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	// 初始化配置
	confFilename := conf.FileArg()
	masterConf, err := conf.ReadMasterConf(confFilename)
	if err != nil {
		fmt.Printf("Init conf error: %s\n", err)
		os.Exit(1)
	}

	// 初始化日志
	err = logs.InitLoggers(masterConf.LogErrorFile)
	if err != nil {
		fmt.Printf("Init log error: %s\n", err)
		os.Exit(1)
	}

	// 初始化环境
	initEnv()
	logs.Info.Printf("Initialized env.")

	// 初始化ETCD连接
	err = master.InitJobManager(masterConf)
	if err != nil {
		logs.Error.Printf("Init Etcd client error: %s", err)
		os.Exit(1)
	}
	logs.Info.Printf("Connected to etcd server: %s",
		masterConf.EtcdEndPoints)

	// 初始化HttpApiServer
	err = master.InitApiServer(masterConf)
	if err != nil {
		logs.Error.Printf("Init HttpServer error: %s", err)
		os.Exit(1)
	}
	logs.Info.Printf("Initialized HttpServer: %s",
		common.GetHost(masterConf.HttpAddress, masterConf.HttpPort))

	// 启动HttpApiServer
	err = master.ApiServerStartListen()
	if err != nil {
		logs.Error.Printf("HttpServer start listen error: %s", err)
		os.Exit(1)
	}
	logs.Info.Printf("HttpServer started listening...")

	time.Sleep(time.Hour)
}
