/*
 * @Descripttion: HTTP协议传输类
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 21:58:52
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-05 22:24:50
 */
 
 package common

//Job 定时任务
type Job struct{
	Name string		`json:"name"`		//任务名
	Command string	`json:"command"`	//shell命令
	CronExpr string	`json:"cronExpr"`	//cron表达式
}