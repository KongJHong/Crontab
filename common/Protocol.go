/*
 * @Descripttion: HTTP协议传输类
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 21:58:52
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-06 09:40:54
 */

 package common

import (
	"encoding/json"
)

//Job 定时任务
type Job struct{
	Name string		`json:"name"`		//任务名
	Command string	`json:"command"`	//shell命令
	CronExpr string	`json:"cronExpr"`	//cron表达式
}

//Response HTTP接口应答
type Response struct{
	Error int  			`json:"error"`	//0表示正常
	Msg string 			`json:"msg"`
	Data interface{} 	`json:"data"`
}

//BuildResponse 应答方法
func BuildResponse(errno int,msg string,data interface{}) (resp []byte,err error) {
	
	//1.定义一个response
	var (
		response Response
	)


	response.Error = errno
	response.Msg = msg
	response.Data = data

	//2.序列化json
	resp,err = json.Marshal(response)
	return 
}