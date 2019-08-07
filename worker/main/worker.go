 package main

 import (
	 "time"
	 "flag"
	 "fmt"
	 "runtime"
	 "Crontab/worker"
 )
 
 
 var (
	 confFile string 	//配置文件路径
 )
 
 //initArgs 解析命令行参数
 func initArgs(){
	 //worker -config ./worker.json
	 //p:入参  name:参数名字 value:默认值 usage: worker -h时能看到的介绍
	 flag.StringVar(&confFile, "config", "./worker.json", "指定worker.json")
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
	 if err = worker.InitConfig(confFile);err != nil{
		 goto ERR
	 }

	 //服务注册
	 if err = worker.InitRegister();err != nil{
		 goto ERR
	 }

	 //启动日志协程
	 if err = worker.InitLogSink();err != nil{
		 goto ERR
	 }

	 //启动执行器
	 if err = worker.InitExecutor();err != nil{
		 goto ERR
	 }

	 //启动调度器
	 if err = worker.InitScheduler();err != nil{
		 goto ERR
	 }
 
	//初始化任务管理器
	if err = worker.InitJobMgr();err != nil{
		goto ERR
	}
	 
	 //正常退出
	 for{
		 time.Sleep(1 * time.Second)
	 }
	 
	 
 ERR:
	 fmt.Println(err)
 }