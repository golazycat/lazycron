package master

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/conf"
	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/common/protocol"
)

const (
	HttpParamParseErrorNo      = 1
	HttpParamJsonDecodeErrorNo = 2
	JobManagerErrorNo          = 3
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
// job: {
//     "name": "任务名称",
//     "command": "任务命令",
//     "cronExpr": "cron表达式",
// }
//
// Return:
// 	  当是覆盖保存时，data为被替代的Job；否则，data为null
//
func handleJobSave(w http.ResponseWriter, r *http.Request) {

	postJob := parseFormAndGet(w, r, "job")
	if postJob == "" {
		return
	}

	var job protocol.Job
	if err := json.Unmarshal([]byte(postJob), &job); err != nil {
		protocol.HttpFail(w, HttpParamJsonDecodeErrorNo,
			fmt.Sprintf("job decode error: %s", postJob), nil)
		return
	}

	oldJob, err := JobManager.SaveJob(&job)
	if err != nil {
		jobManagerError(w, "save", err)
		return
	}

	protocol.HttpSuccess(w, oldJob)
}

// 删除任务
// Method: POST
// Request Body:
//    name: 要删除的job的名称
// Return:
//    若删除成功，data保存被删除的job，删除失败data为null
func handleJobDelete(w http.ResponseWriter, r *http.Request) {

	jobName := parseFormAndGet(w, r, "name")
	if jobName == "" {
		return
	}

	delJob, err := JobManager.DeleteJob(jobName)
	if err != nil {
		jobManagerError(w, "del", err)
		return
	}

	protocol.HttpSuccess(w, delJob)
}

// 列出所有任务
// Method: POST
// Response Body:
//     data会保存所有任务列表，如果没有任务，data保存空列表
func handleJobList(w http.ResponseWriter, _ *http.Request) {

	jobs, err := JobManager.ListJobs()
	if err != nil {
		return
	}

	protocol.HttpSuccess(w, jobs)
}

// 强制杀死某个任务
// 参数同删除任务
// 这个Api仅仅会发送命令，不关心命令的执行结果，命令会由worker执行
// 因此当返回成功时，仅表示命令成功被发送，不表示任务真的被杀死
func handleJobKill(w http.ResponseWriter, r *http.Request) {

	jobName := parseFormAndGet(w, r, "name")
	if jobName == "" {
		return
	}

	err := JobManager.KillJob(jobName)
	if err != nil {
		jobManagerError(w, "kill", err)
		return
	}
	protocol.HttpSuccess(w, nil)
}

// 返回lazycron的HTTP管理服务器.
// 启动该服务器来对整个架构进行管理.
func InitApiServer(conf *conf.MasterConf) error {

	mux := http.NewServeMux()

	// job api
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/del", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)

	// static web root
	staticDir := http.Dir(conf.StaticWebRoot)
	staticHandler := http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

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

// 辅助函数，解析表单并返回某个key的value
// 如果解析失败或者key不存在，会自动将错误json写入responseWriter并返回""
func parseFormAndGet(w http.ResponseWriter, r *http.Request, paramName string) string {

	if err := r.ParseForm(); err != nil {
		protocol.HttpFail(w, HttpParamParseErrorNo,
			"parse form error", nil)
		return ""
	}

	val := r.PostForm.Get(paramName)
	if val == "" {
		protocol.HttpFail(w, HttpParamParseErrorNo,
			fmt.Sprintf("require param %s", paramName), nil)
		return ""

	}
	return val
}

// JobManager错误通用返回
func jobManagerError(w http.ResponseWriter, op string, err error) {
	protocol.HttpFail(w, JobManagerErrorNo,
		fmt.Sprintf("job %s error: %s", op, err), nil)

}
