package common

// 通用常量定义
const (
	// ETCD中用到的各种前缀
	JobKeyPrefix    = "/lazycron/jobs/"
	JobKillPrefix   = "/lazycron/kill/"
	JobLockPrefix   = "/lazycron/lock/"
	JobWorkerPrefix = "/lazycron/worker/"

	// mongodb中用到的常量
	MongodbDatabase   = "lazycron"
	MongodbCollection = "job_log"
)
