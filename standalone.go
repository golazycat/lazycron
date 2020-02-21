package main

import (
	"fmt"

	"github.com/golazycat/lazycron/common"
	"github.com/golazycat/lazycron/master"
	"github.com/golazycat/lazycron/worker"
)

func main() {

	fmt.Println("Starting master..")
	master.Start(true)

	fmt.Println("Starting worker..")
	worker.Start(true)

	common.LoopForever()

}
