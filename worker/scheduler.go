package worker

import (
	"context"
	"os"
	"time"

	job2 "github.com/golazycat/lazycron/common/job"

	"github.com/gorhill/cronexpr"

	"github.com/golazycat/lazycron/common/logs"
	"github.com/golazycat/lazycron/common/protocol"
)

const timeFormat = "2006-01-02 15:04:05"

// Job调度计划。Job的调度执行是依据cron表达式控制的，cron表达式决定了job多久执行一次
// 因此需要这个结构体储存解析好的cron表达式对象，并保存job下一次执行的时间
// 调度器依据这个计划来在规定得时间执行job，并更新下一次执行时间
// 每个计划对象和job对象是一对一的关系
type JobSchedulePlan struct {
	Job      *protocol.Job
	Expr     *cronexpr.Expression
	NextTime time.Time
}

// Job执行信息。Job在执行前，需要创建这个对象传给Executor，里面存了Job执行前的一些参数
// 其中，Job存了执行job的信息，PlanTime表示job的计划执行时间，RealTime表示job的真正执行时间
// 这二者可能有差异是因为机器等各种不可控因素，job的真正执行时间和计划时间不匹配
// 这些情况都需要让Executor知晓，因此保存在这个结构体中
// 另外，一个执行中的job可能随时会被kill掉，这个操作需要用到context的cancelContext
// Executor在执行的时候需要注册这个context，随后Scheduler就可以随时通过cancelFunc来中途
// 中断允许中的job了
type JobExecuteInfo struct {
	Job        *protocol.Job
	PlanTime   time.Time
	RealTime   time.Time
	CancelCtx  context.Context
	CancelFunc context.CancelFunc
}

// Job执行结果。Job在由Executor执行完成后，Executor会创建这个对象并返回给Scheduler(通过channel)
// 这里面保存了job执行的各种信息，包括执行的输出，是否成功，执行时间
// Scheduler收到这个对象会把对应的任务从执行表中删除，从而可以等待下一次执行
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output      []byte
	Err         error
	StartTime   time.Time
	EndTime     time.Time
}

// 创建调度计划，这个过程会解析job对象中的cron表达式，如果解析失败，会返回错误
func CreateJobSchedulerPlan(job *protocol.Job) (*JobSchedulePlan, error) {

	expr, err := cronexpr.Parse(job.CronExpr)
	if err != nil {
		return nil, err
	}
	plan := JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return &plan, nil
}

// 创建Job执行信息，执行信息中的PlanTime由计划plan的NextTime决定，而RealTime由当前时间决定
func CreateJobExecuteInfo(plan *JobSchedulePlan) *JobExecuteInfo {

	cancelCtx, cancelFunc := context.WithCancel(context.TODO())
	return &JobExecuteInfo{
		Job:        plan.Job,
		PlanTime:   plan.NextTime,
		RealTime:   time.Now(),
		CancelCtx:  cancelCtx,
		CancelFunc: cancelFunc,
	}
}

// 调度器结构
// jobEventChan: 用于从JobWorker那里获取job的变化事件从而进行处理
// jobResultChan: 用于从Executor那里获取job的执行结果
// planTable: 保存当前所有job的计划，里面有重要的下一次执行时间
// jobExecuteTable: 保存当前正在执行的所有jobs，key是jobName，value是job plan
// logJob: 在job调度执行的过程中是否输出日志，注意如果设为true，日志将会很长
type SchedulerBody struct {
	jobEventChan    chan *protocol.JobEvent
	jobResultChan   chan *JobExecuteResult
	planTable       map[string]*JobSchedulePlan
	jobExecuteTable map[string]*JobExecuteInfo

	logJob bool
}

// 开始调度，调用这个函数，调度器开始工作
// 调度器需要从JobWorker那里获取job，在特定的时机将job发给Executor执行
// 因此，调用这个函数前JobWorker和Executor必须初始化完成
func (scheduler *SchedulerBody) BeginScheduling() {

	CheckSchedulerInit()

	go scheduler.scheduleLoop()
}

// 处理一个job事件。job事件由JobWorker负责监听并发给Scheduler
// 如果事件是更新，则需要为这个job创建新的计划并加到计划表里；如果是删除，则需要从计划表里删除这个job
// 如果事件是强杀，
func (scheduler *SchedulerBody) handleJobEvent(jobEvent *protocol.JobEvent) {

	switch jobEvent.EventType {
	case protocol.JobEventUpdate:
		plan, err := CreateJobSchedulerPlan(jobEvent.Job)
		if err != nil {
			if scheduler.logJob {
				logs.Warn.Printf("invalid cron expr '%s', the job named %s"+
					" will not be executed", jobEvent.Job.CronExpr, jobEvent.Job.Name)
			}
			return
		}
		scheduler.planTable[jobEvent.Job.Name] = plan

	case protocol.JobEventDelete:
		if _, exists := scheduler.planTable[jobEvent.Job.Name]; exists {
			delete(scheduler.planTable, jobEvent.Job.Name)
		}

	case protocol.JobEventKill:
		if jobExecuteInfo, exist :=
			scheduler.jobExecuteTable[jobEvent.Job.Name]; exist {

			jobExecuteInfo.CancelFunc()

			if scheduler.logJob {
				logs.Info.Printf("killed job: %s", jobEvent.Job.Name)
			}
		}
	}

}

// 处理一个job运行结果。运行结果是由Executor返回给Scheduler的
// 当job执行完毕，需要及时从执行中任务列表中删除这个job
func (scheduler *SchedulerBody) handleJobResult(jobResult *JobExecuteResult) {

	jobName := jobResult.ExecuteInfo.Job.Name

	// 从执行任务中删除该job
	delete(scheduler.jobExecuteTable, jobName)

	// 生成job log，加到db
	if jobResult.Err != LockOccupiedError {

		job := jobResult.ExecuteInfo.Job
		jobLog := protocol.JobLog{
			JobName:          job.Name,
			Command:          job.Command,
			Output:           string(jobResult.Output),
			PlanTime:         jobResult.ExecuteInfo.PlanTime.Unix(),
			ScheduleTime:     jobResult.ExecuteInfo.RealTime.Unix(),
			ExecuteStartTime: jobResult.StartTime.Unix(),
			ExecuteEndTime:   jobResult.EndTime.Unix(),
		}

		if jobResult.Err != nil {
			jobLog.Err = jobResult.Err.Error()
		} else {
			jobLog.Err = ""
		}

		job2.Logger.Insert(&jobLog)
	}
}

// 浏览计划表中的所有job，执行其中需要执行的job，并更新下一次执行时间
// Scheduler需要不停地检查计划表，以确保job尽量在计划时间执行，因此这个函数会被反复调用
// 为了节省CPU利用率，函数会返回计划表中从当前开始下一次job过期的最近时间(如果当前计划表为空，则返回1秒)
// 这样在调度时，可以随时修改该函数的调用时机为该值，以减少甚至杜绝空转(扫描了一次计划表，却没有任何job执行)的次数
func (scheduler *SchedulerBody) scanPlanTable() time.Duration {

	now := time.Now()
	var near *time.Time = nil

	for _, plan := range scheduler.planTable {
		if plan.NextTime.Before(now) || plan.NextTime.Equal(now) {
			scheduler.executeJob(plan)
			plan.NextTime = plan.Expr.Next(now)
		}

		if near == nil || plan.NextTime.Before(*near) {
			near = &plan.NextTime
		}
	}

	if near != nil {
		return (*near).Sub(now)
	}

	return time.Second
}

// 调度执行的主要循环
// 调度执行需要做3个事情：
//     1. 从JobWorker那里获取job事件，并进行处理
//     2. 从Executor那里获取job执行结果，并处理
//     3. 定时扫描计划表执行job
// 其中前二者通过channel进行获取，并传给对应的handle函数
// 3中，需要不停地重置定时器，以减少空转(见scanPlanTable函数)
func (scheduler *SchedulerBody) scheduleLoop() {

	next := scheduler.scanPlanTable()

	// 调度定时器
	timer := time.NewTimer(next)

	for {
		select {

		case jobEvent := <-scheduler.jobEventChan:
			if scheduler.logJob {
				logs.Info.Printf("receive job event: %d %+v",
					jobEvent.EventType, jobEvent.Job)
			}
			scheduler.handleJobEvent(jobEvent)

		case <-timer.C:
			next = scheduler.scanPlanTable()
			timer.Reset(next)

		case jobResult := <-scheduler.jobResultChan:
			scheduler.handleJobResult(jobResult)
		}
	}

}

// 执行job，这个函数会构造执行信息，并发送给Executor来实现对job的执行
// 同时，会将job加到执行表中。
// 如果job已经存在于执行表中了，说明该job的上一次执行还没有结束，将不会
// 执行操作，以防止job的重复并发执行
func (scheduler *SchedulerBody) executeJob(plan *JobSchedulePlan) {

	if _, exists := scheduler.jobExecuteTable[plan.Job.Name]; !exists {
		executeInfo := CreateJobExecuteInfo(plan)
		scheduler.jobExecuteTable[plan.Job.Name] = executeInfo
		Executor.Execute(executeInfo)

		if scheduler.logJob {
			logs.Info.Printf("execute job: plan=%s real=%s job=%s",
				executeInfo.PlanTime.Format(timeFormat),
				executeInfo.RealTime.Format(timeFormat), plan.Job)
		}
	} else {
		if scheduler.logJob {
			logs.Warn.Printf("execute job failed, job "+
				"is still executing... name=%s", plan.Job.Name)
		}
	}
}

// 提交一个job事件给Scheduler
// 由JobWorker调用，这是暴露给外部的接口，用来告诉Scheduler job的变化
// 具体的调度过程不需要外部关心
func (scheduler *SchedulerBody) PushEvent(jobEvent *protocol.JobEvent) {

	CheckSchedulerInit()

	go func() {
		scheduler.jobEventChan <- jobEvent
	}()
}

// 提交一个job执行结果给Scheduler
// 由Executor调用，这是暴露给外部的接口，用来告诉Scheduler job的执行结果
// 具体的调度过程不需要外部关心
func (scheduler *SchedulerBody) PushJobResult(result *JobExecuteResult) {

	CheckSchedulerInit()

	go func() {
		scheduler.jobResultChan <- result
	}()
}

var (
	// 单例模式
	Scheduler SchedulerBody
	isSInit   = false
)

// Scheduler初始化器
type SchedulerInitializer struct {
	LogJob bool
}

// 初始化Scheduler
func (s SchedulerInitializer) Init() error {
	Scheduler = SchedulerBody{
		jobEventChan:    make(chan *protocol.JobEvent),
		planTable:       make(map[string]*JobSchedulePlan),
		jobExecuteTable: make(map[string]*JobExecuteInfo),
		jobResultChan:   make(chan *JobExecuteResult),
		logJob:          s.LogJob,
	}
	isSInit = true
	return nil
}

// 检查Scheduler是否被初始化
func CheckSchedulerInit() {
	if !isSInit {
		logs.Error.Printf("Job Worker Not init!")
		os.Exit(1)
	}
}
