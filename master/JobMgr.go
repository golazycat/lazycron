package master

import (
	"context"
	"encoding/json"
	"os"

	"github.com/golazycat/lazycron/common/protocol"

	"github.com/golazycat/lazycron/common"

	"github.com/golazycat/lazycron/common/conf"

	"github.com/coreos/etcd/clientv3"
	"github.com/golazycat/lazycron/common/logs"
)

const jobKeyPrefix = "/lazycron/jobs/"

// 任务管理器结构
// 保存etcd-cli的对象，以操作etcd来管理job
type JobManagerBody struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

// 保存任务
// 这个函数会把一个新的Job发布到etcd中去，其它worker可以接收这个Job
// etcd中Job的Key将会是job.Name，Value是job序列化的结果
// 如果这个KV之前已经存在了(发生了替换行为)，则该函数会将旧的job反序列化后返回
// 注意如果旧的job反序列化失败，函数不会产生异常
func (jobManager *JobManagerBody) SaveJob(job *protocol.Job) (*protocol.Job, error) {

	CheckJobManagerInit()

	jobKey := jobKeyPrefix + job.Name
	jobValue, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}

	// 保存到etcd
	putResponse, err := JobManager.kv.Put(context.TODO(), jobKey,
		string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	// 如果是更新操作，需要把旧的值返回出去
	if putResponse.PrevKv != nil {
		var oldJob protocol.Job
		if err = json.Unmarshal(putResponse.PrevKv.Value, &oldJob); err != nil {
			// 这里旧值反序列化失败对新值的保存没有影响，因此不做出错情况处理
			return nil, nil
		}
		return &oldJob, nil
	}

	return nil, nil
}

var (
	// 全局Job Manager
	JobManager JobManagerBody
	// Job Manager是否初始化?
	isJMInit = false
)

func CheckJobManagerInit() {
	if !isJMInit {
		logs.Error.Printf("Job Manager Not init!")
		os.Exit(1)
	}
}

func InitJobManager(conf *conf.MasterConf) error {

	config := clientv3.Config{
		Endpoints:   conf.EtcdEndPoints,
		DialTimeout: common.IntSecond(conf.EtcdDialTimeout),
	}

	client, err := clientv3.New(config)
	if err != nil {
		return err
	}

	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	JobManager = JobManagerBody{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	isJMInit = true

	return nil
}
