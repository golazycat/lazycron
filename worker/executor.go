package worker

import (
	"context"
	"os"
	"os/exec"
	"time"

	"github.com/golazycat/lazycron/common/logs"
)

// 执行器结构体，执行器用于从Scheduler那里获取需要执行的job并执行
// 执行job完毕后将执行结果返回给Scheduler，是一个中间件
type ExecutorBody struct {
}

// 执行指定的job，并将执行结果返回给Scheduler
// 执行过程会异步进行
// 注意，在执行前，需要尝试获取这个job的分布式锁，如果获取失败，说明
// 有其他的worker正在执行这个job，则会跳过这个job的执行
func (executor *ExecutorBody) Execute(info *JobExecuteInfo) {

	CheckExecutorInit()

	go func() {

		result := JobExecuteResult{
			ExecuteInfo: info,
			StartTime:   time.Now(),
		}
		jobLock := CreateJobLock(info.Job.Name, &JobWorker.Connector)
		defer jobLock.UnLock()

		if err := jobLock.Lock(); err != nil {
			// 抢占锁失败，错误退出
			result.EndTime = time.Now()
			result.Err = LockOccupiedError

		} else {

			// 抢占分布式锁需要花时间，因此这里重置开始时间
			result.StartTime = time.Now()

			cmd := exec.CommandContext(context.TODO(),
				"/bin/bash", "-c", info.Job.Command)

			output, err := cmd.CombinedOutput()
			if output == nil {
				output = make([]byte, 0)
			}

			result.EndTime = time.Now()
			result.Output = output
			result.Err = err

		}

		Scheduler.PushJobResult(&result)
	}()

}

var (
	Executor *ExecutorBody
	isEInit  = false
)

type ExecutorInitializer struct {
}

func (e ExecutorInitializer) Init() error {

	Executor = &ExecutorBody{}
	isEInit = true
	return nil
}

func CheckExecutorInit() {
	if !isEInit {
		logs.Error.Printf("Job Worker Not init!")
		os.Exit(1)
	}
}
