/*
 * @Descripttion: HTTP初始化，RAII配置初始路由，把任务保存到ETCD中
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 21:05:30
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-05 22:15:48
 */

 package master

import (
	"strconv"
	"time"
	"net"
	"net/http"
//	"Crontab/common"
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
func handleJobSave(w http.ResponseWriter,r *http.Request){
	//任务保存到ETCD中
	
}

//InitAPIServer HTTP服务器初始化服务函数
func InitAPIServer() (err error){

	var (
		mux *http.ServeMux
		listener net.Listener
		httpServer *http.Server
	)

	

	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)


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