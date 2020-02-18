package main

import (
	"os"

	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/baseconf"
	"github.com/golazycat/lazycron/common/baseinit"
	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/worker"
	"github.com/golazycat/lazycron/worker/conf"
)

func main() {

	confFilename := baseconf.FileArg()
	workerConf := conf.ReadWorkerConf(confFilename)

	baseinit.Init(baseinit.LoggersInitializer{
		ErrorFilePath: ""}, "log")

	baseinit.Init(worker.JobWorkerInitializer{
		Conf: workerConf}, "job worker")

	logs.Info.Printf("use conf: %+v", workerConf)

	logs.Info.Printf("begin watching jobs...")
	err := worker.JobWorker.BeginWatchJobs()
	if err != nil {
		logs.Error.Printf("watch job error: %s", err)
		os.Exit(1)
	}

	common.LoopForever()

}
