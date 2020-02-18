package common

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/golazycat/lazycron/common/protocol"
)

func IntSecond(t int) time.Duration {
	return time.Duration(t) * time.Second
}

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
