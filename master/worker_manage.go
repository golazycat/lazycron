package master

import (
	"context"
	"os"

	"github.com/coreos/etcd/clientv3"
	"github.com/golazycat/lazycron/common"

	"github.com/golazycat/lazycron/common/baseconf"
	"github.com/golazycat/lazycron/common/etcd"
	"github.com/golazycat/lazycron/common/logs"
)

// 保存所有注册的worker的信息
// ID表示worker的唯一标识
type Worker struct {
	ID string `json:"id"`
}

// Worker管理器结构体
// Worker是注册到etcd的，所以需要etcd连接
type WorkerManagerBody struct {
	etcd.Connector
}

// 取得目前在线的所有worker，将从etcd进行获取
func (workerManager *WorkerManagerBody) GetWorkers() ([]*Worker, error) {

	CheckWorkerManagerInit()

	getResponse, err := workerManager.Kv.Get(context.TODO(),
		common.JobWorkerPrefix, clientv3.WithPrefix())

	if err != nil {
		return nil, err
	}

	workers := make([]*Worker, 0)

	for _, kv := range getResponse.Kvs {
		id := common.GetIDFromWorker(kv)
		worker := Worker{ID: id}
		workers = append(workers, &worker)
	}

	return workers, nil
}

// WorkerManager 初始化器
type WorkerManagerInitializer struct {
	Conf baseconf.EtcdConf
}

var (
	WorkerManager WorkerManagerBody
	isWMInit      = false
)

// WorkerManager 初始化
func (w WorkerManagerInitializer) Init() error {

	conn, err := etcd.CreateConnect(&w.Conf)
	if err != nil {
		return err
	}

	WorkerManager.Connector = *conn

	isWMInit = true

	return nil
}

func CheckWorkerManagerInit() {
	if !isWMInit {
		logs.Error.Printf("worker manager not init!")
		os.Exit(1)
	}
}
