package worker

import (
	"os"

	"github.com/golazycat/lazycron/common/baseconf"
	"github.com/golazycat/lazycron/common/baseinit"
	"github.com/golazycat/lazycron/common/joblog"
	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/worker/conf"
)

func Start(simple bool) {

	var confFilename = ""
	if !simple {
		confFilename = baseconf.FileArg()
	}
	workerConf := conf.ReadWorkerConf(confFilename)

	baseinit.Init(baseinit.LoggersInitializer{
		ErrorFilePath: ""}, "log")

	baseinit.Init(RegisterInitializer{
		Conf: workerConf.EtcdConf}, "register")

	baseinit.Init(joblog.LoggerInitializer{
		Conf: workerConf.MongoConf}, "job log")
	joblog.Logger.BeginListening()

	baseinit.Init(JobWorkerInitializer{
		Conf: workerConf}, "job worker")

	baseinit.Init(ExecutorInitializer{}, "executor")

	baseinit.Init(SchedulerInitializer{
		LogJob: workerConf.LogJob}, "scheduler")

	logs.Info.Printf("use conf: %+v", workerConf)

	err := JobWorker.BeginWatchJobs()
	logs.Info.Printf("begin watching jobs...")
	if err != nil {
		logs.Error.Printf("watch job error: %s", err)
		os.Exit(1)
	}

	Scheduler.BeginScheduling()
	logs.Info.Printf("begin scheduling...")

}
