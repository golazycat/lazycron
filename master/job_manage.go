package master

import (
	"context"
	"encoding/json"
	"os"

	"github.com/golazycat/lazycron/master/conf"

	"github.com/golazycat/lazycron/common/etcd"

	"github.com/golazycat/lazycron/common/protocol"

	"github.com/golazycat/lazycron/common"

	"github.com/coreos/etcd/clientv3"
	"github.com/golazycat/lazycron/common/logs"
)

// 任务管理器结构
// 保存etcd-cli的对象，以操作etcd来管理job
type JobManagerBody struct {
	etcd.Connector
}

// 这个函数会把一个新的Job发布到etcd中去，其它worker可以接收这个Job
// etcd中Job的Key将会是job.Name，Value是job序列化的结果
// 如果这个KV之前已经存在了(发生了替换行为)，则该函数会将旧的job反序列化后返回
// 注意如果旧的job反序列化失败，函数不会产生异常
func (jobManager *JobManagerBody) SaveJob(job *protocol.Job) (*protocol.Job, error) {

	CheckJobManagerInit()

	jobKey := common.JobKeyPrefix + job.Name
	jobValue, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	// 保存到etcd
	putResponse, err := JobManager.Kv.Put(context.TODO(), jobKey,
		string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	// 如果是更新操作，需要把旧的值返回出去
	if putResponse.PrevKv != nil {
		return common.GetJobFromKv(putResponse.PrevKv), nil
	}

	return nil, nil
}

// 这个函数会将指定name的Job从etcd中删除
// 如果删除成功，会返回删除的Job对象指针；指定name的job不存在，则会返回nil
func (jobManager *JobManagerBody) DeleteJob(name string) (*protocol.Job, error) {

	CheckJobManagerInit()

	jobKey := common.JobKeyPrefix + name

	delResponse, err := jobManager.Kv.Delete(context.TODO(),
		jobKey, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}
	if len(delResponse.PrevKvs) != 0 {
		// 删除操作针对单一job进行
		return common.GetJobFromKv(delResponse.PrevKvs[0]), nil
	}
	return nil, nil
}

// 列出所有任务，从etcd中获取job目录下所有的任务
func (jobManager *JobManagerBody) ListJobs() ([]*protocol.Job, error) {

	CheckJobManagerInit()

	listResponse, err := jobManager.Kv.Get(context.TODO(),
		common.JobKeyPrefix, clientv3.WithPrefix())

	if err != nil {
		return nil, err
	}

	jobs := make([]*protocol.Job, 0)
	for _, kv := range listResponse.Kvs {
		if job := common.GetJobFromKv(kv); job != nil {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}

// 发出杀死任务命令给workers
// 这个操作会在etcd的KillJobPrefix目录下新加需要kill的jobName
// 这个kv只会存在1秒的时间，1秒后会自动到期被删除，worker只需要监听到这个变化即可
// 其它worker收到这个信号之后会去执行kill，随后master就不管了
func (jobManager JobManagerBody) KillJob(name string) error {

	CheckJobManagerInit()

	killKey := common.JobKillPrefix + name

	leaseGrantResponse, err :=
		jobManager.Lease.Grant(context.TODO(), 1)
	if err != nil {
		return err
	}
	leaseId := leaseGrantResponse.ID

	// 这里不关心kill put操作的返回结果，执行成功即可
	_, err = jobManager.Kv.Put(context.TODO(), killKey,
		"", clientv3.WithLease(leaseId))
	if err != nil {
		return err
	}
	return nil
}

var (
	// 全局Job Manager
	JobManager JobManagerBody
	// Job Manager是否初始化?
	isJMInit = false
)

// 检查JobManager是否被初始化，即是否调用成功过InitJobManager函数
// 如果没有调用，产生Fatal并终止程序
func CheckJobManagerInit() {
	if !isJMInit {
		logs.Error.Printf("Job Manager Not init!")
		os.Exit(1)
	}
}

// JobManager初始化
type JobManagerInitializer struct {
	Conf *conf.MasterConf
}

// 初始化JobManager，在使用JobManager之前，必须调用这个函数，否则使用JobManager的任何函数
// 会产生Fatal错误，直接退出程序。
// 这个函数会尝试去连接etcd服务器，如果连接失败，会返回错误。
func (j JobManagerInitializer) Init() error {

	conn, err := etcd.CreateConnect(&j.Conf.EtcdConf)
	if err != nil {
		return err
	}

	JobManager = JobManagerBody{}
	JobManager.Connector = *conn

	isJMInit = true

	return nil

}
