package conf

import "github.com/golazycat/lazycron/common/baseconf"

// Master配置结构
type MasterConf struct {
	baseconf.EtcdConf
	baseconf.RunConf
	HttpAddress      string `json:"http.addr"`
	HttpPort         int    `json:"http.port"`
	HttpReadTimeout  int    `json:"http.read_timeout"`
	HttpWriteTimeout int    `js￿on:"http.write_timeout"`
	StaticWebRoot    string `json:"http.static_path"`
	LogErrorFile     string `json:"log.error_path"`
}

// Master默认配置
func (c *MasterConf) SetDefault() {

	c.EtcdConf.SetDefault()
	c.RunConf.SetDefault()

	c.HttpAddress = ""
	c.HttpPort = 8070
	c.HttpReadTimeout = 5
	c.HttpWriteTimeout = 5
	c.StaticWebRoot = "./static"
	c.LogErrorFile = ""

}

// 读取配置文件，返回Master配置对象指针
// 如果读取有问题或filename为""，会使用默认配置
func ReadMasterConf(filename string) *MasterConf {

	var masterConf MasterConf
	baseconf.ReadConf(filename, &masterConf)

	return &masterConf
}
