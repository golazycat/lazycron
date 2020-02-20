package job

import (
	"context"
	"os"

	"github.com/golazycat/lazycron/common/protocol"

	"github.com/golazycat/lazycron/common/baseconf"
	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/common/mongo"
)

// 使用mongodb来记录job log
// log的格式为JobLog结构体
type LoggerBody struct {
	mongo.Connector
}

// 新加一个log
func (logger *LoggerBody) Insert(jobLog *protocol.JobLog) error {

	CheckLoggerInit()

	_, err := logger.Collection.InsertOne(context.TODO(), jobLog)
	if err != nil {
		return err
	}

	return nil
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
	isLInit = true

	return nil
}

func CheckLoggerInit() {
	if !isLInit {
		logs.Error.Printf("job logger not init")
		os.Exit(1)
	}
}
