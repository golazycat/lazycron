package common

const (
	JobKeyPrefix  = "/lazycron/jobs/"
	JobKillPrefix = "/lazycron/kill/"
	JobLockPrefix = "/lazycron/lock/"

	MongodbDatabase   = "lazycron"
	MongodbCollection = "job_log"
)
