package protocol

import (
	"encoding/json"
	"net/http"
)

// Job事件类型枚举
const (
	JobEventDelete = iota
	JobEventUpdate
	JobEventKill
)

// 定时任务结构
// 保存任务所需要的数据，由Master分配给Worker执行
type Job struct {
	// 任务名称
	Name string `json:"name"`
	// 任务命令
	Command string `json:"command"`
	// Cron 表达式
	CronExpr string `json:"cron_expr"`
}

// Job事件结构体，保存了事件类型和产生事件对应的job指针
type JobEvent struct {
	EventType int
	Job       *Job
}

type JobLog struct {
	JobName          string `json:"job_name" bson:"job_name"`
	Command          string `json:"command" bson:"command"`
	Err              string `json:"err" bson:"err"`
	Output           string `json:"output" bson:"output"`
	PlanTime         int64  `json:"plan_time" bson:"plan_time"`
	ScheduleTime     int64  `json:"schedule_time" bson:"schedule_time"`
	ExecuteStartTime int64  `json:"exec_start_time" bson:"exec_start_time"`
	ExecuteEndTime   int64  `json:"exec_end_time" bson:"exec_end_time"`
}

// Http API返回的所有数据都遵循这个结构
type HttpResponse struct {
	// 出错码，正常为0
	ErrorNo int `json:"errno"`
	// 请求结果信息
	Message string `json:"message"`
	// 请求返回数据
	Data interface{} `json:"data"`
}

// 创建一个job事件对象
// eventType建议使用JobEventXxx枚举类型
func CreateJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}

}

// Http请求成功，errorno固定为0，message固定为"ok"
func HttpSuccess(w http.ResponseWriter, data interface{}) {
	_, _ = w.Write(createHttpResponseBytes(0, "ok", data))
}

// Http请求失败，需要传入errno和message
func HttpFail(w http.ResponseWriter, errNo int, msg string, data interface{}) {
	_, _ = w.Write(createHttpResponseBytes(errNo, msg, data))
}

// 以bytes的形式获取HttpResponse结构体
func createHttpResponseBytes(errNo int, msg string, data interface{}) []byte {

	resp := HttpResponse{
		ErrorNo: errNo,
		Message: msg,
		Data:    data,
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		return []byte("json decode error")
	}
	return respJson
}
