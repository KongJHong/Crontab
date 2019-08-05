/*
 * @Descripttion: 管理任务增删改查（和ETCD直接交互）
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 22:00:58
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-05 22:14:01
 */

 package master

import (
	"time"
	"github.com/coreos/etcd/clientv3"
)
 
 
//任务管理器
 type JobMgr struct{
	 client *clientv3.Client
	 kv clientv3.KV
	 lease clientv3.Lease
 }

 var (
	 //单例
	 G_jobMgr *JobMgr
 )


 //InitJobMgr 初始化管理器
 func InitJobMgr() (err error){
	 
	var (
		config clientv3.Config
		client *clientv3.Client
		kv   clientv3.KV
		lease clientv3.Lease
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints :G_config.EtcdEndpoints,	//集群地址
		DialTimeout : time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,	//连接超时
	}

	//建立连接
	if client,err = clientv3.New(config);err != nil{
		return
	}


	//得到KV和lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//赋值单例
	G_jobMgr = &JobMgr{
		client:client,
		kv:kv,
		lease:lease,
	}

	return 
 }