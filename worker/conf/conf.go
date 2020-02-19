package conf

import "github.com/golazycat/lazycron/common/baseconf"

type WorkerConf struct {
	baseconf.EtcdConf
	baseconf.RunConf
	LogJob bool `json:"log_job"`
}

func (conf *WorkerConf) SetDefault() {
	conf.EtcdConf.SetDefault()
	conf.RunConf.SetDefault()

	conf.LogJob = true
}

func ReadWorkerConf(filename string) *WorkerConf {
	var conf WorkerConf
	baseconf.ReadConf(filename, &conf)
	return &conf
}
