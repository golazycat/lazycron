package conf

import (
	"fmt"
	"testing"
)

func TestReadMasterConf(t *testing.T) {

	masterConf := ReadMasterConf("master.json")
	fmt.Printf("%+v\n", masterConf)

	masterConf = ReadMasterConf("")
	fmt.Printf("%+v\n", masterConf)

}
