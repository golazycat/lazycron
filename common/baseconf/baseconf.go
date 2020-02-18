package baseconf

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"runtime"
)

// 配置接口，所有配置结构体都需要实现这个接口
// 实现接口后即可调用ReadConf函数来实现配置的自动赋值
type Config interface {

	// 配置对象需要通过它来设置配置的默认值
	// 在读取配置的时候，可能没有提供指定配置或其他原因导致配置读取失败
	// 这个时候会调用该函数，在函数中需要将配置赋值为默认配置
	SetDefault()
}

// 读取配置，写入v中。
// 配置文件需要是一个json数据，这个函数会将读取的json文件反序列化后写入v中
// 如果希望使用默认配置，则filename设置为""即可。
// 如果在读取配置的过程中产生任何错误，都会使用默认配置。
// 关于如何编写默认配置，见Config接口。
func ReadConf(filename string, v Config) {

	if filename == "" {
		v.SetDefault()
		return
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		v.SetDefault()
		return
	}

	err = json.Unmarshal(content, v)
	if err != nil {
		v.SetDefault()
	}
}

// 从命令行参数取得conf参数，表示配置文件路径。默认为""
func FileArg() string {

	var (
		confFilePath string
	)

	flag.StringVar(&confFilePath, "conf", "",
		"Configuration file path, format see document."+
			" Not setting will use the default configuration.")
	flag.Parse()
	return confFilePath
}

// etcd配置结构，保存了连接etcd所需要的配置项
type EtcdConf struct {
	EtcdDialTimeout int      `json:"etcd.dial_timeout"`
	EtcdEndPoints   []string `json:"etcd.endpoints"`
}

// etcd默认配置
func (conf *EtcdConf) SetDefault() {

	conf.EtcdEndPoints = []string{"localhost:2379"}
	conf.EtcdDialTimeout = 5

}

// 运行配置，保存运行程序的一些参数信息
type RunConf struct {
	NThread int `json:"run.n_thread"`
}

// 运行默认配置
func (conf *RunConf) SetDefault() {
	conf.NThread = runtime.NumCPU()
}
