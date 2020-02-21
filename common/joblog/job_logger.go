package joblog

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/golazycat/lazycron/common/protocol"

	"github.com/golazycat/lazycron/common/baseconf"
	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/common/mongo"
)

// 使用mongodb来记录job log
// log的格式为JobLog结构体
type LoggerBody struct {
	mongo.Connector
	logChan   chan *protocol.JobLog
	batchSize int
}

type LogBatch struct {
	logs []interface{}
}

type LogFilter struct {
	JobName string `bson:"job_name"`
}

type SortLogByStartTime struct {
	SortOrder int `bson:"exec_start_time"`
}

func (logger *LoggerBody) BeginListening() {
	go logger.listeningLogs()
}

func (logger *LoggerBody) listeningLogs() {

	CheckLoggerInit()

	timer := time.NewTicker(time.Second)

	var logBatch *LogBatch = nil
	flash := func(l int) {
		if logBatch != nil && len(logBatch.logs) > l {
			logger.insertBatch(logBatch)
			logBatch = nil
		}
	}

	for {

		select {
		case jobLog := <-logger.logChan:
			if logBatch == nil {
				logBatch = &LogBatch{}
			}

			logBatch.logs = append(logBatch.logs, jobLog)
			flash(logger.batchSize - 1)

		case <-timer.C:
			flash(0)
		}

		time.Sleep(100 * time.Millisecond)
	}

}

func (logger *LoggerBody) insertBatch(batch *LogBatch) {
	_, _ = logger.Collection.InsertMany(context.TODO(), batch.logs)
}

// 新加一个log
func (logger *LoggerBody) Insert(jobLog *protocol.JobLog) {

	CheckLoggerInit()

	go func() {
		logger.logChan <- jobLog
	}()

}

// 根据job name来查找所有的log
func (logger *LoggerBody) FindByJobLogName(
	jobName string, skip int64, limit int64) ([]*protocol.JobLog, error) {

	CheckLoggerInit()

	filter := &LogFilter{JobName: jobName}

	// 根据任务开始时间对log进行排序
	logSort := SortLogByStartTime{SortOrder: -1}

	cursor, err := logger.Collection.Find(context.TODO(), filter,
		options.Find().SetSort(logSort).SetSkip(skip).SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	result := make([]*protocol.JobLog, 0)
	for cursor.Next(context.TODO()) {
		jobLob := protocol.JobLog{}
		if err = cursor.Decode(&jobLob); err != nil {
			continue
		}
		result = append(result, &jobLob)
	}

	return result, nil
}

var (
	Logger  LoggerBody
	isLInit = false
)

// job logger初始化
type LoggerInitializer struct {
	Conf baseconf.MongoConf
}

func (l LoggerInitializer) Init() error {

	conn, err := mongo.CreateConnect(&l.Conf)
	if err != nil {
		return err
	}

	Logger.Connector = *conn
	Logger.logChan = make(chan *protocol.JobLog)
	Logger.batchSize = l.Conf.WriteBatchSize
	isLInit = true

	return nil
}

func CheckLoggerInit() {
	if !isLInit {
		logs.Error.Printf("job logger not init")
		os.Exit(1)
	}
}
