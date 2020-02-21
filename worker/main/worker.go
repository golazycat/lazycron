package main

import (
	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/worker"
)

func main() {

	worker.Start(false)
	common.LoopForever()

}
