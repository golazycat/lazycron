package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/golazycat/lazycron/common/protocol"
)

// 将int转换为Duration->n秒
func IntSecond(t int) time.Duration {
	return time.Duration(t) * time.Second
}

// 将地址和端口转换为标准的host地址输出
func GetHost(addr string, port int) string {
	return fmt.Sprintf("%s:%d", addr, port)
}

// 从KV对中的Value获取Job对象
// 这个过程需要取出Value，按照json进行解析，反序列化后返回
// 如果解析失败，会返回nil
func GetJobFromKv(kv *mvccpb.KeyValue) *protocol.Job {

	var job protocol.Job
	if err := json.Unmarshal(kv.Value, &job); err != nil {
		return nil
	}

	return &job
}

// 从KV对中的Key取得job名称
func GetJobNameFromKv(kv *mvccpb.KeyValue) string {
	return strings.TrimPrefix(string(kv.Key), JobKeyPrefix)
}

// 从KV kill中的key取得job名称
func GetJobNameFromKill(kv *mvccpb.KeyValue) string {
	return strings.TrimPrefix(string(kv.Key), JobKillPrefix)
}

// 让程序永远运行下去
func LoopForever() {
	for {
		time.Sleep(time.Second)
	}
}

// 终端字体颜色支持
const (
	ColorFontGreen = iota
	ColorFontRed
	ColorFontYellow
)

var colorMap = map[int]int{
	ColorFontGreen: 32, ColorFontRed: 31, ColorFontYellow: 33,
}

// 返回有颜色的字体
func ColorString(s string, color int) string {
	return fmt.Sprintf("\033[%d;1m%s\033[0m", colorMap[color], s)
}
