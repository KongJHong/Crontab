/*
 * @Descripttion:  
 * @version: 1
 * @Author: KongJHong
 * @Date: 2019-08-05 21:02:05
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-08 09:45:52
 */


package main

import (
	"time"
	"flag"
	"fmt"
	"runtime"
	"Crontab/master"
)


var (
	confFile string 	//配置文件路径
)

//initArgs 解析命令行参数
func initArgs(){
	//master -config ./master.json
	//p:入参  name:参数名字 value:默认值 usage: master -h时能看到的介绍
	flag.StringVar(&confFile, "config", "./master.json", "指定master.json")
	flag.Parse()//正式解析
}

//initEnv 初始化线程数量
func initEnv(){
	runtime.GOMAXPROCS(runtime.NumCPU())
}


func init(){
	//初始化命令行参数
	initArgs()

	//初始化线程,要发挥它的多核优势，就必须限制它的线程数量等于它的核心数量
	initEnv()
}


func main(){

	var (
		err error
	)

	//加载配置
	if err = master.InitConfig(confFile);err != nil{
		goto ERR
	}

	//初始化服务发现
	if err = master.InitWorkerMgr();err != nil{
		goto ERR 
	}

	//日志管理器
	if err = master.InitLogMgr();err != nil{
		goto ERR
	}

	//任务管理器
	if err = master.InitJobMgr();err != nil{
		goto ERR
	}

	//启动API HTTP服务
	if err = master.InitAPIServer();err != nil{	//HTTP在协程中跑的
		goto ERR
	}

	
	//正常退出
	for{
		time.Sleep(1 * time.Second)
	}
	
	
ERR:
	fmt.Println(err)
}