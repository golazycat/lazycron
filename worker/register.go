package worker

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"

	"github.com/google/uuid"

	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/baseconf"
	"github.com/golazycat/lazycron/common/etcd"
	"github.com/golazycat/lazycron/common/logs"
)

// worker注册器结构体
type RegisterBody struct {
	etcd.Connector

	localIp string
}

// 保持worker在线，worker使用etcd通知master自己在线，这个过程是通过保存key实现的
// 这个函数会在etcd创建当前worker的在线标识，随后使用租约机制让其一直保持在线
// 只要注册器在线且etcd正常，该key就会一直存在，master可以通过监听这个key来知晓这个worker的在线情况
// 当worker离线时，因为租约机制，key会在一定时间后自动被取消，master就可以及时知道worker的离线
func (register *RegisterBody) keepOnline() {

	regKey := common.JobWorkerPrefix + register.localIp

	var (
		keepChan   <-chan *clientv3.LeaseKeepAliveResponse
		cancelFunc context.CancelFunc
	)
	for {
		leaseResponse, err := register.Lease.Grant(context.TODO(), 10)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		keepChan, err = register.Lease.KeepAlive(
			context.TODO(), leaseResponse.ID)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		cancelCtx, cancelFunc := context.WithCancel(context.TODO())

		_, err = register.Kv.Put(cancelCtx,
			regKey, "", clientv3.WithLease(leaseResponse.ID))
		if err != nil {
			time.Sleep(time.Second)
			cancelFunc()
			continue
		}

		break

	}

	for {
		keepAlive := <-keepChan
		if keepAlive == nil { // 租约失效，重新keep
			cancelFunc()
			register.keepOnline()
		}
	}

}

var (
	Register RegisterBody
)

// 注册器初始化
type RegisterInitializer struct {
	Conf baseconf.EtcdConf
}

// 初始化注册器，这个过程会让worker一直尝试保持在线
func (r RegisterInitializer) Init() error {

	ip, err := common.GetLocalIp()
	if err != nil {
		ip = uuid.New().String()
		logs.Warn.Printf("No local ip found in"+
			" this machine! Will use UUID %s to registe worker.", ip)
	}

	conn, err := etcd.CreateConnect(&r.Conf)
	if err != nil {
		return err
	}

	Register.Connector = *conn
	Register.localIp = ip

	go Register.keepOnline()

	return nil
}
