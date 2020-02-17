package master

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/golazycat/lazycron/common/protocol"

	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/conf"
	"github.com/golazycat/lazycron/common/logs"
)

const (
	HttpParamParseErrorNo      = 1
	HttpParamJsonDecodeErrorNo = 2
	JobManagerErrorNo          = 3
	OtherServerErrorNo         = 4
)

var (
	// 全局Http server
	gHttpServer *ApiServer
	// http server是否初始化
	isASInit = false
	// 全局Http listener
	httpListener net.Listener
)

// 对外暴露的HTTP接口.
// 外部需要通过HTTP接口来管理整个lazycron.
// 包括任务的发布、查询、强杀，日志的查询等功能.
// 所以提供的Api在此实现
type ApiServer struct {
	httpServer *http.Server
}

// 保存任务
// Method: POST
// Request Body:
// "job": {
//     "name": "任务名称",
//     "command": "任务命令",
//     "cronExpr": "cron表达式",
// }
//
// Return:
// 	  当是覆盖保存时，data为被替代的Job；否则，data为null
//
func handleJobSave(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		protocol.HttpFail(w, HttpParamParseErrorNo,
			"require job param", nil)
		return
	}

	postJob := r.PostForm.Get("job")

	var job protocol.Job
	if err := json.Unmarshal([]byte(postJob), &job); err != nil {
		protocol.HttpFail(w, HttpParamJsonDecodeErrorNo,
			fmt.Sprintf("job decode error: %s", postJob), nil)
		return
	}

	oldJob, err := JobManager.SaveJob(&job)
	if err != nil {
		protocol.HttpFail(w, JobManagerErrorNo,
			fmt.Sprintf("job save error: %s", err), nil)
		return
	}

	protocol.HttpSuccess(w, oldJob)
}

// 返回lazycron的HTTP管理服务器.
// 启动该服务器来对整个架构进行管理.
func InitApiServer(conf *conf.MasterConf) error {

	mux := http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)

	var err error
	httpListener, err = net.Listen("tcp",
		common.GetHost(conf.HttpAddress, conf.HttpPort))
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		ReadTimeout:  common.IntSecond(conf.HttpReadTimeout),
		WriteTimeout: common.IntSecond(conf.HttpWriteTimeout),
		Handler:      mux,
	}

	gHttpServer = &ApiServer{httpServer: httpServer}
	isASInit = true

	return nil
}

// 调用该函数即可启动一个goroutine来监听API Http请求
// 在调用前需要调用InitApiServer来初始化Http服务器，否则会产生错误
func ApiServerStartListen() error {
	if !isASInit {
		return errors.New("ApiServer is not initialized")
	}
	go func() {

		err := gHttpServer.httpServer.Serve(httpListener)
		if err != nil {
			logs.Error.Printf("Fatal Error: Http Sevrer start fail!, error: %v", err)
			os.Exit(1)
		}
	}()
	return nil
}
