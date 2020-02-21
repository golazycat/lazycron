package main

import (
	"os"

	"github.com/golazycat/lazycron/common/joblog"

	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/baseinit"

	"github.com/golazycat/lazycron/common/baseconf"

	"github.com/golazycat/lazycron/master/conf"

	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/master"
)

func main() {

	// 初始化配置
	confFilename := baseconf.FileArg()
	masterConf := conf.ReadMasterConf(confFilename)

	// 初始化日志
	baseinit.Init(baseinit.LoggersInitializer{
		ErrorFilePath: masterConf.LogErrorFile}, "log")

	baseinit.Init(joblog.LoggerInitializer{
		Conf: masterConf.MongoConf}, "job log")

	baseinit.Init(master.WorkerManagerInitializer{
		Conf: masterConf.EtcdConf}, "worker manager")

	// 初始化环境
	baseinit.Init(baseinit.RunInitializer{
		RunConf: &masterConf.RunConf}, "runtime")

	// 初始化JobManager
	baseinit.Init(master.JobManagerInitializer{
		Conf: masterConf}, "job manager")

	// 初始化HttpApiServer
	baseinit.Init(master.ApiServerInitializer{
		Conf: masterConf}, "http api")

	// 启动HttpApiServer
	err := master.ApiServerStartListen()
	if err != nil {
		logs.Error.Printf("HttpServer start listen error: %s", err)
		os.Exit(1)
	}
	logs.Info.Printf("HttpServer started listening...")

	// 打印下配置
	logs.Info.Printf("use conf: %+v", masterConf)

	common.LoopForever()
}
