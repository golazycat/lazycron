package etcd

import (
	"github.com/coreos/etcd/clientv3"
	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/conf"
)

// etcd连接对象，保存了操作etcd所需要的各种对象
type Connector struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

// 新建一个etcd连接，返回连接对象
// 参数中包括了连接etcd所需要的各种配置
func CreateConnect(conf *conf.EtcdConf) (*Connector, error) {

	config := clientv3.Config{
		Endpoints:   conf.EtcdEndPoints,
		DialTimeout: common.IntSecond(conf.EtcdDialTimeout),
	}

	client, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)

	return &Connector{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}, nil
}
