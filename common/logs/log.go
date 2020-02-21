package logs

import (
	"io"
	"log"
	"os"

	"github.com/golazycat/lazycron/common"
)

var (
	// INFO级别日志
	Info *log.Logger
	// WARN级别日志
	Warn *log.Logger
	// ERROR级别日志
	Error *log.Logger
)

// 初始化日志器
// errPath参数可以指定单独的错误输出路径，如果设为""，则错误只输出到Stderr
func InitLoggers(errorPath string) error {

	var errWriter io.Writer

	if errorPath != "" {
		errFile, err := os.OpenFile(errorPath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		errWriter = io.MultiWriter(os.Stderr, errFile)
	} else {
		errWriter = os.Stderr
	}

	Error = log.New(errWriter, common.ColorString("[ERR] ", common.ColorFontRed), log.LstdFlags)
	Warn = log.New(os.Stdout, common.ColorString("[WAR] ", common.ColorFontYellow), log.LstdFlags)
	Info = log.New(os.Stdout, common.ColorString("[INF] ", common.ColorFontGreen), log.LstdFlags)

	return nil
}
