package conf

import "github.com/golazycat/lazycron/common/baseconf"

type WorkerConf struct {
	baseconf.EtcdConf
	baseconf.RunConf
}

func (conf *WorkerConf) SetDefault() {
	conf.EtcdConf.SetDefault()
	conf.RunConf.SetDefault()

}

func ReadWorkerConf(filename string) *WorkerConf {
	var conf WorkerConf
	baseconf.ReadConf(filename, &conf)
	return &conf
}
