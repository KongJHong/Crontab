/*
 * @Descripttion: HTTP协议传输类
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 21:58:52
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-06 21:38:24
 */

 package common

import (
	"github.com/gorhill/cronexpr"
	"strings"
	"encoding/json"
	"time"
)

//Job 定时任务
type Job struct{
	Name string		`json:"name"`		//任务名
	Command string	`json:"command"`	//shell命令
	CronExpr string	`json:"cronExpr"`	//cron表达式
}

//JobSchedulePlan 任务调度计划
type JobSchedulePlan struct{
	Job *Job //调度的任务信息
	Expr *cronexpr.Expression //解析好的cronexpr表达式
	NextTime time.Time		//下次调度时间
}

//Response HTTP接口应答
type Response struct{
	Errno int  			`json:"errno"`	//0表示正常
	Msg string 			`json:"msg"`
	Data interface{} 	`json:"data"`
}

//JobEvent 变化事件
type JobEvent struct{
	EventType int //SAVE,DELETE
	Job *Job
}

//BuildResponse 应答方法
func BuildResponse(errno int,msg string,data interface{}) (resp []byte,err error) {
	
	//1.定义一个response
	var (
		response Response
	)


	response.Errno = errno
	response.Msg = msg
	response.Data = data

	//2.序列化json
	resp,err = json.Marshal(response)
	return 
}

//UnpackageJob 反序列化job
func UnpackageJob(value []byte) (ret *Job,err error){

	var (
		job *Job
	)

	job = &Job{}

	if err = json.Unmarshal(value, job);err != nil{
		return
	}

	ret = job
	return 
}

//ExtractJobName 从ETCD的key中提取任务名
// 	/cron/jobs/job10 抹掉 /cron/jobs/
func ExtractJobName(jobKey string) (string){
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

//BuildJobEvent 任务变化事件，有两种，1)更新任务，2)删除任务
func BuildJobEvent(eventType int,job *Job)(jobEvent *JobEvent){
	return &JobEvent{
		EventType:eventType,
		Job:job,
	}
}


//BuildJobSchedulePlan 构造执行计划
func BuildJobSchedulePlan(job *Job)(jobSchedulePlan *JobSchedulePlan,err error){
	var (
		expr *cronexpr.Expression
	)

	//解析job的cron表达式
	if expr,err = cronexpr.Parse(job.CronExpr);err != nil{
		return 
	}

	//生成任务调度计划对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:job,
		Expr:expr,
		NextTime:expr.Next(time.Now()),
	}

	return 
}