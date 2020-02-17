package conf

import (
	"fmt"
	"testing"
)

func TestReadMasterConf(t *testing.T) {

	// 默认配置
	conf, _ := ReadMasterConf("")
	fmt.Printf("%+v\n", conf)

	// 读取配置
	conf, _ = ReadMasterConf("master.json")
	fmt.Printf("%+v\n", conf)

}
