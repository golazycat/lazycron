package worker

import (
	"context"
	"os"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/etcd"
	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/common/protocol"
	"github.com/golazycat/lazycron/worker/conf"
)

// 任务执行结构
// 保存etcd-cli的对象，以操作etcd来管理job
type JobWorkerBody struct {
	etcd.Connector
}

// 处理etcd某个key变化函数
type watchHandleFunc func(*clientv3.Event)

// 调用该函数，首先会遍历所有的job，并将这些job保存为job update事件提交给scheduler
// 随后，从遍历的最后一个job的revision开始，调用etcd的watcher监听job的变化
// 当job产生变化，该函数会创建一个job变化事件，并将该事件提交给scheduler执行
func (jobWorker *JobWorkerBody) BeginWatchJobs() error {

	CheckJobWorkerInit()

	getResponse, err := jobWorker.Kv.Get(context.TODO(),
		common.JobKeyPrefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, kv := range getResponse.Kvs {
		if job := common.GetJobFromKv(kv); job != nil {

			jobEvent := protocol.CreateJobEvent(protocol.JobEventUpdate, job)
			Scheduler.PushEvent(jobEvent)
		}
	}

	go jobWorker.keepWatchJobs(getResponse.Header.Revision)
	go jobWorker.keepWatchKills()

	return nil
}

// 处理一个新watch到的event，构建JobEvent对象，发给Scheduler调度
func (jobWorker *JobWorkerBody) handleJobWatchEvent(event *clientv3.Event) {

	var jobEvent *protocol.JobEvent = nil

	switch event.Type {
	case mvccpb.PUT:
		var job *protocol.Job
		if job = common.GetJobFromKv(event.Kv); job == nil {
			return
		}

		jobEvent = protocol.CreateJobEvent(protocol.JobEventUpdate, job)

	case mvccpb.DELETE:
		joName := common.GetJobNameFromKv(event.Kv)
		job := protocol.Job{Name: joName}

		jobEvent = protocol.CreateJobEvent(protocol.JobEventDelete, &job)
	}

	if jobEvent == nil {
		// 未知的事件不予处理
		return
	}

	Scheduler.PushEvent(jobEvent)

}

// 处理一个新的kill请求，转换为jobEvent，发送给Scheduler处理
func (jobWorker *JobWorkerBody) handleKillWatchEvent(event *clientv3.Event) {

	if event.Type == mvccpb.PUT {

		jobName := common.GetJobNameFromKill(event.Kv)
		jobEvent := protocol.CreateJobEvent(protocol.JobEventKill,
			&protocol.Job{Name: jobName})
		Scheduler.PushEvent(jobEvent)
	}

}

// 当初始jobs读取完成后，会获得最后的一个revision，该函数从最后的revision的下一个开始进行
// 监听。当job发生变化，会根据变化创建对应的事件，并将事件提交给scheduler执行
// initLastRevision指定了初始化的最后一个revision
func (jobWorker *JobWorkerBody) keepWatchJobs(initRevision int64) {
	jobWorker.keepWatch(common.JobKeyPrefix, initRevision, jobWorker.handleJobWatchEvent)
}

// 持续监听kill请求的变化，当监听到master发送的kill请求时，进行处理
func (jobWorker *JobWorkerBody) keepWatchKills() {
	jobWorker.keepWatch(common.JobKillPrefix, -1, jobWorker.handleKillWatchEvent)
}

// 辅助函数，在etcd监听某个key的变化，需要传入对应的变化处理函数
// initLastRevision表示从哪个revision开始监听，如果从默认的revision开始监听，传-1
func (jobWorker *JobWorkerBody) keepWatch(watchKey string, initLastRevision int64, handleFunc watchHandleFunc) {

	watcher := clientv3.NewWatcher(jobWorker.Client)

	var watchChan clientv3.WatchChan
	if initLastRevision != -1 {
		watchStartRevision := initLastRevision + 1
		watchChan = watcher.Watch(context.TODO(), watchKey,
			clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
	} else {
		watchChan = watcher.Watch(context.TODO(), watchKey, clientv3.WithPrefix())
	}

	// 处理监听事件
	for watchResponse := range watchChan {
		for _, watchEvent := range watchResponse.Events {
			handleFunc(watchEvent)
		}
	}
}

var (
	JobWorker JobWorkerBody
	isJWInit  = false
)

type JobWorkerInitializer struct {
	Conf *conf.WorkerConf
}

// 初始化JobWorker，在使用JobWorker之前，必须调用这个函数，否则使用JobWorker的任何函数
// 会产生Fatal错误，直接退出程序。
// 这个函数会尝试去连接etcd服务器，如果连接失败，会返回错误。
func (j JobWorkerInitializer) Init() error {

	conn, err := etcd.CreateConnect(&j.Conf.EtcdConf)
	if err != nil {
		return err
	}

	JobWorker = JobWorkerBody{}
	JobWorker.Connector = *conn

	isJWInit = true
	return nil

}

// 检查JobWorker是否被初始化，即是否调用成功过InitJobWorker函数
// 如果没有调用，产生Fatal并终止程序
func CheckJobWorkerInit() {
	if !isJWInit {
		logs.Error.Printf("Job Worker Not init!")
		os.Exit(1)
	}
}
