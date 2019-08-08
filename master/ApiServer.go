/*
 * @Descripttion: HTTP初始化，RAII配置初始路由，把任务保存到ETCD中
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 21:05:30
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-08 09:44:17
 */

 package master

import (
	"encoding/json"
	"strconv"
	"time"
	"net"
	"net/http"
	"Crontab/common"
)

//APIServer 任务的HTTP接口
type APIServer struct{
	httpServer *http.Server
}


var (
	//单例对象
	G_apiServer *APIServer
)


//handleJobSave 保存任务接口
//POST job={"name":"job1","command":"echo hello","cronExpr":"* * * * *"}
func handleJobSave(resp http.ResponseWriter,req *http.Request){
	
	var (
		err error
		postJob string
		job common.Job
		oldJob *common.Job
		bytes []byte
	)
	//任务保存到ETCD中
	//1.解析POST表单
	if err = req.ParseForm();err != nil{
		goto ERR
	}

	//2.取表单中的Job字段
	postJob = req.PostForm.Get("job")

	//3.反序列化job
	if err = json.Unmarshal([]byte(postJob), &job);err != nil{
		goto ERR
	}


	//持久化保存ETCD，所以要传给JobMgr,JobMgr保存到ETCD中
	//4.保存到etcd
	if oldJob,err = G_jobMgr.SaveJob(&job);err != nil{
		goto ERR
	}


	//5.返回正常应答{"error":0,"msg":"","data":{...}}
	if bytes,err = common.BuildResponse(0, "suceess", oldJob);err == nil{
		resp.Write(bytes)
	}

	

	return 
ERR:
	//6，返回异常应答
	if bytes,err = common.BuildResponse(-1, err.Error(), nil);err == nil{
		resp.Write(bytes)
	}
}

//handleJobDelete 删除任务接口
// POST /job/delete name=job1
func handleJobDelete(resp http.ResponseWriter,req *http.Request){

	var (
		err error
		name string
		oldJob *common.Job
		bytes []byte
	)

	//POST: a=1&b=2&c=3
	if err = req.ParseForm();err != nil{
		goto ERR
	}

	//删除的任务名
	name = req.PostForm.Get("name")
	
	//删除任务
	if oldJob,err = G_jobMgr.deleteJob(name);err != nil{
		goto ERR
	}

	//正常应答
	if bytes,err = common.BuildResponse(0, "success", oldJob);err == nil{
		resp.Write(bytes)
	}

	return 
ERR:
	//错误应答
	if bytes,err = common.BuildResponse(-1, err.Error(), nil);err == nil{
		resp.Write(bytes)
	}
}

//handleJobList 列举所有crontab任务
func handleJobList(resp http.ResponseWriter,req *http.Request){
	var (
		jobList []*common.Job
		err error
		bytes []byte
	)
	
	if jobList,err = G_jobMgr.ListJobs();err != nil{
		goto ERR 
	}
	
	//返回正常应答
	if bytes,err = common.BuildResponse(0, "success", jobList);err == nil{
		resp.Write(bytes)
	}

	return 
ERR:
	//错误应答
	if bytes,err = common.BuildResponse(-1, err.Error(), nil);err == nil{
		resp.Write(bytes)
	}
}


//handleJobKill 强制杀死某个任务
// POST /job/kill name=job1
func handleJobKill(resp http.ResponseWriter,req *http.Request){
	var (
		err error
		name string
		bytes []byte
	)

	//解析POST表单
	if err = req.ParseForm();err != nil{
		goto ERR
	}

	//要杀死的任务名
	name = req.PostForm.Get("name")
	
	//杀死任务
	if err = G_jobMgr.KillJob(name);err != nil{
		goto ERR
	}

	//返回正常应答
	if bytes,err = common.BuildResponse(0, "success", nil);err == nil{
		resp.Write(bytes)
	}

	return 
ERR:
	//错误应答
	if bytes,err = common.BuildResponse(-1, err.Error(), nil);err == nil{
		resp.Write(bytes)
	}
}


//handleJobKill 查询任务日志
func handleJobLog(resp http.ResponseWriter,req *http.Request){
	
	var (
		err error
		name string	//任务名字
		skipParam string//从第几条开始
		limitParam string//限制返回多少条
		skip int
		limit int
		logArr []*common.JobLog
		bytes []byte
	)
	
	//解析GET参数
	if err = req.ParseForm();err != nil{
		goto ERR 
	}

	//获取请求参数 /cron/job?name=job10&skip=0&limit=10
	name = req.Form.Get("name")
	limitParam = req.Form.Get("limit")
	skipParam = req.Form.Get("skip")

	if skip,err = strconv.Atoi(skipParam);err != nil{
		skip = 0
	}

	if limit,err = strconv.Atoi(limitParam);err != nil{
		limit = 20
	}

	if logArr,err = G_logMgr.ListLog(name, skip, limit);err != nil{
		goto ERR
	}

	//返回正常应答
	if bytes,err = common.BuildResponse(0, "success", logArr);err == nil{
		resp.Write(bytes)
	}

	return 
ERR:
	//错误应答
	if bytes,err = common.BuildResponse(-1, err.Error(), nil);err == nil{
		resp.Write(bytes)
	}
}

//handleWorkerList 获取健康worker节点列表
func handleWorkerList(resp http.ResponseWriter,req *http.Request){
	var (
		workerArr []string
		err error
		bytes []byte
	)

	if workerArr,err = G_workerMgr.ListWorkers();err != nil{
		goto ERR
	}

	//返回正常应答
	if bytes,err = common.BuildResponse(0, "success", workerArr);err == nil{
		resp.Write(bytes)
	}

	return 
ERR:
	//错误应答
	if bytes,err = common.BuildResponse(-1, err.Error(), nil);err == nil{
		resp.Write(bytes)
	}

}


//InitAPIServer HTTP服务器初始化服务函数,开启HTTP服务器
func InitAPIServer() (err error){

	var (
		mux *http.ServeMux
		listener net.Listener
		httpServer *http.Server
		staticDir http.Dir		//静态文件根目录
		staticHandler http.Handler //静态文件的Http回调
	)

	

	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)		//增-改
	mux.HandleFunc("/job/delete",handleJobDelete)	//删
	mux.HandleFunc("/job/list",handleJobList)		//查
	mux.HandleFunc("/job/kill",handleJobKill)		//杀死任务
	mux.HandleFunc("/job/log", handleJobLog)		//任务日志查询
	mux.HandleFunc("/worker/list", handleWorkerList)//获取worker列表

	//静态文件目录 7.9
	staticDir = http.Dir(G_config.WebRoot)				//设置静态文件根目录
	staticHandler = http.FileServer(staticDir)		//系统设置
	mux.Handle("/",http.StripPrefix("/", staticHandler)) //http.StripPrefix是抹掉/ 然后去./webroot找index.html

	

	
	//启动TCP监听
	if listener,err = net.Listen("tcp", ":"+strconv.Itoa(G_config.APIPort));err != nil{
		return 
	}

	//创建一个HTTP服务
	httpServer = &http.Server{
		ReadTimeout:time.Duration(G_config.APIReadTimeout) * time.Millisecond,		//读超时 5s
		WriteTimeout:time.Duration(G_config.APIWriteTimeout) * time.Millisecond,	//写超时 5s
		Handler:mux,
	}

	//赋值单例
	G_apiServer = &APIServer{
		httpServer : httpServer,
	}

	//启动了服务端
	go httpServer.Serve(listener)

	return 
}