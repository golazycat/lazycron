package master

import (
	"os"

	"github.com/golazycat/lazycron/common/baseconf"
	"github.com/golazycat/lazycron/common/baseinit"
	"github.com/golazycat/lazycron/common/joblog"
	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/master/conf"
)

func Start(simple bool) {

	// 初始化配置
	var confFilename string
	if !simple {
		confFilename = baseconf.FileArg()
	} else {
		confFilename = ""
	}
	masterConf := conf.ReadMasterConf(confFilename)

	// 初始化日志
	baseinit.Init(baseinit.LoggersInitializer{
		ErrorFilePath: masterConf.LogErrorFile}, "log")

	baseinit.Init(joblog.LoggerInitializer{
		Conf: masterConf.MongoConf}, "job log")

	baseinit.Init(WorkerManagerInitializer{
		Conf: masterConf.EtcdConf}, "worker manager")

	// 初始化环境
	baseinit.Init(baseinit.RunInitializer{
		RunConf: &masterConf.RunConf}, "runtime")

	// 初始化JobManager
	baseinit.Init(JobManagerInitializer{
		Conf: masterConf}, "job manager")

	// 初始化HttpApiServer
	baseinit.Init(ApiServerInitializer{
		Conf: masterConf}, "http api")

	// 启动HttpApiServer
	err := ApiServerStartListen()
	if err != nil {
		logs.Error.Printf("HttpServer start listen error: %s", err)
		os.Exit(1)
	}
	logs.Info.Printf("HttpServer started listening...")

	// 打印下配置
	logs.Info.Printf("use conf: %+v", masterConf)

}
