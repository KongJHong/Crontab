/*
 * @Descripttion: 任务调度功能类，包含定时器，任务执行
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-06 21:10:09
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-07 17:02:34
 */

package worker

import (
	"fmt"
	"time"
	"Crontab/common"
)

//Scheduler 任务调度
type Scheduler struct{
	jobEventChan chan *common.JobEvent	//etcd任务事件队列
	jobPlanTable map[string]*common.JobSchedulePlan //任务调度计划表，每一个任务下一次的执行时间都在这张表中
	jobExecutingTable map[string]*common.JobExecuteInfo	//任务执行表
	jobResultChan	chan *common.JobExecuteResult		//任务结果队列

}


var (
	G_scheduler *Scheduler
)

//处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent){
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExecuteInfo *common.JobExecuteInfo
		jobExisting bool
		jobExisted bool
		err error
	)
	switch jobEvent.EventType{
	case common.JOB_EVENT_SAVE:		//保存任务事件
		if jobSchedulePlan,err = common.BuildJobSchedulePlan(jobEvent.Job);err != nil{
			return 
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE:	//删除任务事件
		if jobSchedulePlan,jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name];jobExisted{
			delete(scheduler.jobPlanTable,jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL:	//强杀任务事件
		//取消掉Command执行
		if jobExecuteInfo,jobExisting = scheduler.jobExecutingTable[jobEvent.Job.Name];jobExisting{
			fmt.Println("------------------强杀任务:",jobExecuteInfo.Job.Name)
			jobExecuteInfo.CancelFunc() //触发command杀死shell子进程，任务得到退出
		}
	}
}

//TryStartJob 尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan){
	//调度和执行是两件事
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting bool
	)

	//执行的任务可能运行很久，一分钟会调度60次，但是只能执行1次,防止并发！


	//如果任务正在执行，跳过本次调度
	if jobExecuteInfo,jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name];jobExecuting{
		fmt.Println("尚未退出，跳过执行")
		return
	}

	//构建执行状态信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)
	
	//保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	//执行任务
	G_executor.ExecuteJob(jobExecuteInfo)
	fmt.Println("执行任务",jobExecuteInfo.Job.Name,jobExecuteInfo.PlanTime,jobExecuteInfo.RealTime)
}


//TrySchedule 重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule()(scheduleAfter time.Duration){
	
	var (
		jobPlan *common.JobSchedulePlan
		now time.Time
		nearTime *time.Time
	)
	
	//如果任务表为空的话，随便睡眠多久
	if len(scheduler.jobPlanTable) == 0{
		scheduleAfter = 1 * time.Second
		return 
	}

	//获取当前时间
	now = time.Now()

	//1.遍历所有任务
	for _,jobPlan = range scheduler.jobPlanTable{
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now){
			//尝试执行任务
			scheduler.TryStartJob(jobPlan)
			//更新下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)	
		}
		//统计最近一个要过期的任务事件(下一个要执行的任务距离现在又多久)
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime){
			nearTime = &jobPlan.NextTime
		}
	}

	//2.下次调度时间 (最近要调度的任务调度时间-当前时间)
	scheduleAfter = (*nearTime).Sub(now)
	return 
}

//handleJobResult 处理任务结果
func (scheduler *Scheduler)handleJobResult(jobResult *common.JobExecuteResult){
	var (
		jobLog *common.JobLog
	)
	//删除执行状态
	delete(scheduler.jobExecutingTable,jobResult.ExecuteInfo.Job.Name)

	//生成执行日志
	if jobResult.Err != common.ERR_LOCK_ALREADY_REQUIRED{	//对于抢锁失败的错误，应该跳过
		jobLog = &common.JobLog{
			JobName:jobResult.ExecuteInfo.Job.Name,
			Command:jobResult.ExecuteInfo.Job.Command,
			Output:string(jobResult.Output),
			PlanTime:jobResult.ExecuteInfo.PlanTime.UnixNano()/1000/1000,	//精确到毫秒
			ScheduleTime:jobResult.ExecuteInfo.RealTime.UnixNano()/1000/1000,
			StartTime:jobResult.StartTime.UnixNano()/1000/1000,
			EndTime:jobResult.EndTime.UnixNano()/1000/1000,
		}

		if jobResult.Err != nil{	//防止err是空的情况，会报错
			jobLog.Err = jobResult.Err.Error()
		}else{
			jobLog.Err = ""
		}

		//TODO:存储到mongodb,不应该在scheduleLoop中进行插入，因为是和硬盘进行操作，会破坏scheduleLoop的执行逻辑，而是应该另开协程
		G_logSink.Append(jobLog)
	}

	fmt.Println("任务执行完成",jobResult.ExecuteInfo.Job.Name,string(jobResult.Output),jobResult.Err)
}


//调度协程
func (scheduler *Scheduler)scheduleLoop(){
	
	var (
		jobEvent *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult *common.JobExecuteResult
	)

	//初始化一次(1秒)
	scheduleAfter = scheduler.TrySchedule()

	//调度的延时定时器
	scheduleTimer = time.NewTimer(scheduleAfter)
	
	//定时任务 commonJob
	for{
		select {
		case jobEvent = <- scheduler.jobEventChan:	//监听任务变化事件
			//对我们维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C:	//最近的任务到期了
		case jobResult = <- scheduler.jobResultChan: //监听任务结果
			scheduler.handleJobResult(jobResult)
		}
		//调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		//重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}
}

//PushJobEvent 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent){
	scheduler.jobEventChan <- jobEvent
}

//InitScheduler 初始化调度器
func InitScheduler()(err error){
	G_scheduler = &Scheduler{
		jobEventChan:make(chan *common.JobEvent,1000),
		jobPlanTable:make(map[string]*common.JobSchedulePlan),
		jobExecutingTable:make(map[string]*common.JobExecuteInfo),
		jobResultChan:make(chan *common.JobExecuteResult,1000),
	}

	//启动调度协程
	go G_scheduler.scheduleLoop()

	return 
}

//PushJobResult 回传任务执行结果
func (scheduler *Scheduler)PushJobResult(jobResult *common.JobExecuteResult){
	scheduler.jobResultChan <- jobResult
}