package conf

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
)

var (
	ReadConfError   = errors.New("read conf file error")
	JsonDecodeError = errors.New("json decode error")
)

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

// etcd配置结构
type etcdConf struct {
	EtcdDialTimeout int      `json:"etcd.dial_timeout"`
	EtcdEndPoints   []string `json:"etcd.endpoints"`
}

// Master配置结构
type MasterConf struct {
	etcdConf
	HttpAddress      string `json:"http.addr"`
	HttpPort         int    `json:"http.port"`
	HttpReadTimeout  int    `json:"http.read_timeout"`
	HttpWriteTimeout int    `json:"http.write_timeout"`
	LogErrorFile     string `json:"log.error_path"`
}

// 读取Master配置文件，并返回配置对象指针
// 如果参数filename设为""，则会返回默认配置
// 默认配置定义在项目根路径中的master_default.json
func ReadMasterConf(filename string) (*MasterConf, error) {
	if filename == "" {
		return createDefaultMasterConf(), nil
	}
	var masterConf MasterConf
	err := readConf(filename, &masterConf)
	if err != nil {
		return nil, err
	}
	return &masterConf, err
}

// 默认Master配置
func createDefaultMasterConf() *MasterConf {

	etcdConf := etcdConf{
		EtcdDialTimeout: 5,
		EtcdEndPoints:   []string{"localhost:2379"},
	}

	return &MasterConf{
		etcdConf:         etcdConf,
		HttpAddress:      "",
		HttpPort:         8070,
		HttpReadTimeout:  5,
		HttpWriteTimeout: 5,
		LogErrorFile:     "",
	}
}

// 读取配置通用函数
// 读取文件错误返回ReadConfError，Json解析错误返回JsonDecodeError
func readConf(filename string, v interface{}) error {

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return ReadConfError
	}

	err = json.Unmarshal(content, v)
	if err != nil {
		return JsonDecodeError
	}

	return nil
}
