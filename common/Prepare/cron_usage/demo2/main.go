package main

import (
	"github.com/gorhill/cronexpr"
	"time"
	"fmt"
)

//CronJob 代表一个任务
type CronJob struct {
	expr *cronexpr.Expression
	nextTime time.Time	//expr.Next(now)
}

func main(){
	//需要由一个协程，定时检查所有的Cron任务，谁过期了就执行谁

	var (
		cronJob *CronJob
		expr *cronexpr.Expression
		now time.Time
		scheduleTable map[string]*CronJob //key:任务的名字 value就是任务
	)

	scheduleTable = make(map[string]*CronJob)

	//1.定义两个cronjob
	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr : expr,
		nextTime : expr.Next(now),
	}

	//任务1注册到调度表
	scheduleTable["job1"] = cronJob

	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr : expr,
		nextTime : expr.Next(now),
	}

	//任务2注册到调度表
	scheduleTable["job2"] = cronJob


	//启动一个调度协程
	go func(){
		var(
			jobName string
			cronJob *CronJob
			now time.Time
		)
		//定时检查一下任务调度表
		for{
			now = time.Now()
			for jobName,cronJob = range scheduleTable{
				//判断是否过期
				if cronJob.nextTime.Before(now) || cronJob.nextTime.Equal(now){
					//启动一个协程，执行这个任务
					go func(jobName string){
						fmt.Println("执行:",jobName)
					}(jobName)

					//计算下一次的执行时间
					cronJob.nextTime = cronJob.expr.Next(now)
					fmt.Println(jobName,"下次执行时间是:",cronJob.nextTime)
				}
			}

			select{
			case <-time.NewTimer(100 * time.Millisecond).C: //将在100毫秒可读，返回
			}
		}
	}()

	time.Sleep(100 *time.Second)
}