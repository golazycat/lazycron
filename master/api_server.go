package master

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/golazycat/lazycron/common/joblog"

	"github.com/golazycat/lazycron/master/conf"

	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/common/protocol"
)

const (
	HttpParamParseErrorNo = iota + 1
	HttpParamJsonDecodeErrorNo
	JobManagerErrorNo
	JobLogErrorNo
	WorkerManagerErrorNo
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
// Return:
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

// 获取job执行的参数
// Method: POST
// Request Body:
//     name: 要查询的job名称
//     skip: 分页参数
//     limit: 分页参数
// Return:
//     如果查询到日志，data会保存查询到的job日志
func handleJobLog(w http.ResponseWriter, r *http.Request) {

	params := parseFromAndGetMany(w, r, []string{"name", "skip", "limit"})
	if params == nil {
		return
	}

	name := params["name"]
	skip := getIntValueOrDefault(params["skip"], 0)
	limit := getIntValueOrDefault(params["limit"], 20)

	jobs, err := joblog.Logger.FindByJobLogName(name, int64(skip), int64(limit))
	if err != nil {
		protocol.HttpFail(w, JobLogErrorNo,
			fmt.Sprintf("job log error: %s", err), nil)
		return
	}

	protocol.HttpSuccess(w, jobs)
}

// 获取所有在线的workers
// Method: POST
// Return:
//    data为所有在线的worker列表
func handleWorkerList(w http.ResponseWriter, r *http.Request) {

	workers, err := WorkerManager.GetWorkers()
	if err != nil {
		protocol.HttpFail(w, WorkerManagerErrorNo,
			fmt.Sprintf("worker manage error: %s", err), nil)
		return
	}

	protocol.HttpSuccess(w, workers)
}

// ApiServer 初始化器
type ApiServerInitializer struct {
	Conf *conf.MasterConf
}

// 初始化APiServer
func (a ApiServerInitializer) Init() error {

	mux := http.NewServeMux()

	// job api
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/del", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("/job/log", handleJobLog)
	mux.HandleFunc("/worker/list", handleWorkerList)

	// static web root
	staticDir := http.Dir(a.Conf.StaticWebRoot)
	staticHandler := http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

	var err error
	httpListener, err = net.Listen("tcp",
		common.GetHost(a.Conf.HttpAddress, a.Conf.HttpPort))
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		ReadTimeout:  common.IntSecond(a.Conf.HttpReadTimeout),
		WriteTimeout: common.IntSecond(a.Conf.HttpWriteTimeout),
		Handler:      mux,
	}

	gHttpServer = &ApiServer{httpServer: httpServer}
	isASInit = true

	return nil
}

// 调用该函数即可启动一个goroutine来监听API Http请求
// 在调用前需要调用ApiServerInitializer.Init()来初始化Http服务器，否则会产生错误
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

func parseForm(w http.ResponseWriter, r *http.Request) error {

	if err := r.ParseForm(); err != nil {
		protocol.HttpFail(w, HttpParamParseErrorNo,
			"parse form error", nil)
		return err
	}

	return nil
}

// 辅助函数，解析表单并返回某个key的value
// 如果解析失败或者key不存在，会自动将错误json写入responseWriter并返回""
func parseFormAndGet(w http.ResponseWriter, r *http.Request, paramName string) string {

	err := parseForm(w, r)
	if err != nil {
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

func parseFromAndGetMany(w http.ResponseWriter, r *http.Request, paramNames []string) map[string]string {

	err := parseForm(w, r)
	if err != nil {
		return nil
	}

	var params = make(map[string]string, len(paramNames))

	for _, paramName := range paramNames {
		val := parseFormAndGet(w, r, paramName)
		if val == "" {
			return nil
		} else {
			params[paramName] = val
		}
	}

	return params
}

func getIntValueOrDefault(sVal string, valDefault int) int {
	if val, err := strconv.Atoi(sVal); err == nil {
		return val
	}
	return valDefault
}

// JobManager错误通用返回
func jobManagerError(w http.ResponseWriter, op string, err error) {
	protocol.HttpFail(w, JobManagerErrorNo,
		fmt.Sprintf("job %s error: %s", op, err), nil)

}
