package worker

import (
	"context"
	"errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/etcd"
)

var LockOccupiedError = errors.New("job lock is occupied")

// Job分布式锁结构
// 分布式锁由etcd实现，对每个job都有一个锁，在执行job前，应该先尝试获取
// 锁，获取锁成功时，表示没有其它worker在操作这个job，在操作完成之后，需要释放锁
// 分布式锁通过etcd租约+事务来实现
type JobLock struct {
	etcd.Connector

	jobName    string
	cancelFunc context.CancelFunc
	keepChan   <-chan *clientv3.LeaseKeepAliveResponse
	leaseId    clientv3.LeaseID
	isLocked   bool
}

// 创建job分布式锁，jobName表示为哪个job创建的锁，因为锁是通过etcd实现的，所以
// 需要传入可用的etcd连接对象
func CreateJobLock(jobName string, conn *etcd.Connector) *JobLock {
	return &JobLock{
		Connector: *conn,
		jobName:   jobName,
	}
}

// 对job上锁。成功调用这个函数之后，其它worker再对这个job调用该函数时，会返回LockOccupiedError
// 因此，如果在调用的时候返回LockOccupiedError，表示这个job被其它worker占用了，不应该处理之
// 在调用Lock函数后，job处理完成之后应该调用UnLock函数释放锁，否则其它worker将一直无法获取锁
// 注意这个锁并不是阻塞的，获取失败返回error，函数并不会阻塞住
func (jobLock *JobLock) Lock() error {

	cancelCtx, cancelFunc := context.WithCancel(context.TODO())
	jobLock.cancelFunc = cancelFunc

	// 租约，让lock限时存在
	leaseResponse, err := jobLock.Lease.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}
	jobLock.leaseId = leaseResponse.ID

	keepChan, err := jobLock.Lease.KeepAlive(cancelCtx, jobLock.leaseId)
	if err != nil {
		jobLock.cancelLock()
		return err
	}

	// 自动续租
	go func() {
		for {
			keepResponse := <-keepChan
			if keepResponse == nil {
				break
			}
		}
	}()

	// 创建事务
	txn := jobLock.Kv.Txn(context.TODO())

	lockKey := common.JobLockPrefix + jobLock.jobName
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(jobLock.leaseId))).
		Else(clientv3.OpGet(lockKey))

	// 提交事务
	txnResponse, err := txn.Commit()
	if err != nil {
		jobLock.cancelLock()
		return err
	}

	if txnResponse.Succeeded {
		// 锁被占用
		jobLock.cancelLock()
		return LockOccupiedError
	}

	jobLock.isLocked = true

	return nil
}

// 释放锁的具体实现
func (jobLock *JobLock) cancelLock() {
	jobLock.cancelFunc()
	_, _ = jobLock.Lease.Revoke(context.TODO(), jobLock.leaseId)
}

// 释放锁，如果已经获取了锁，调用这个函数会把当前锁释放掉
func (jobLock *JobLock) UnLock() {
	if jobLock.isLocked {
		jobLock.cancelLock()
		jobLock.isLocked = false
	}
}
