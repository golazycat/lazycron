package baseinit

import (
	"os"
	"runtime"

	"github.com/golazycat/lazycron/common"

	"github.com/golazycat/lazycron/common/baseconf"
	"github.com/golazycat/lazycron/common/logs"
)

// 初始化器，用于初始化某个功能
// Init()函数用来执行初始化
type Initializer interface {
	Init() error
}

// 执行初始化操作，就是调用初始化器的Init()函数
// 如果初始化失败，会打印错误信息并退出程序
// 如果初始化成功，会向日志输出信息
func Init(initializer Initializer, opName string) {
	err := initializer.Init()
	if err != nil {
		logs.Error.Printf("%s %s: %s",
			common.ColorString("[x]", common.ColorFontRed), opName, err)
		os.Exit(1)
	}
	logs.Info.Printf("%s %s",
		common.ColorString("[√]", common.ColorFontGreen), opName)
}

// 用于初始化日志
type LoggersInitializer struct {
	ErrorFilePath string
}

// 初始化日志
func (li LoggersInitializer) Init() error {
	err := logs.InitLoggers(li.ErrorFilePath)
	if err != nil {
		return err
	}
	return nil
}

// 用于初始化基本运行环境
type RunInitializer struct {
	RunConf *baseconf.RunConf
}

// 初始化基本运行环境
func (ri RunInitializer) Init() error {
	runtime.GOMAXPROCS(ri.RunConf.NThread)
	return nil
}
