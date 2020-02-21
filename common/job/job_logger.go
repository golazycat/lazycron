package job

import (
	"context"
	"os"
	"time"

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
func (logger *LoggerBody) FindByJobLogName(jobName string) ([]protocol.JobLog, error) {

	CheckLoggerInit()

	return nil, nil
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
