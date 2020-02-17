package protocol

import (
	"encoding/json"
	"net/http"
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

// Http API返回的所有数据都遵循这个结构
type HttpResponse struct {
	// 出错码，正常为0
	ErrorNo int `json:"errno"`
	// 请求结果信息
	Message string `json:"message"`
	// 请求返回数据
	Data interface{} `json:"data"`
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
