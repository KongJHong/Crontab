/*
 * @Descripttion: Worker的JobMgr是监听ETCD的变化（和ETCD直接交互）
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 22:00:58
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-06 21:21:47
 */

 package worker

import (
	"github.com/coreos/etcd/mvcc/mvccpb"
	"Crontab/common"
	"context"
	"time"
	"github.com/coreos/etcd/clientv3"

)
 
 
//JobMgr 任务管理器
 type JobMgr struct{
	 client *clientv3.Client
	 kv clientv3.KV
	 lease clientv3.Lease
	 watcher clientv3.Watcher
 }

 var (
	 //单例
	 G_jobMgr *JobMgr
 )


//watchJobs 监听任务变化
func (jobMgr *JobMgr)watchJobs()(err error){

	var (
		getResp *clientv3.GetResponse
		kvpair *mvccpb.KeyValue
		job *common.Job
		watchStartRevision int64
		watchChan clientv3.WatchChan
		watchResp clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobName string
		jobEvent *common.JobEvent
		
	)
	//1.get一下/cron/jobs/目录下的所有任务，并且获知当前集群的revision
	if getResp,err = jobMgr.kv.Get( context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix());err != nil{
		return 
	}

	//当前有哪些任务,推送到scheduler中
	for _,kvpair = range getResp.Kvs{
		//反序列化json得到job
		if job,err = common.UnpackageJob(kvpair.Value);err == nil{
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			//TODO:把这个job同步给scheduler（调度协程）
			G_scheduler.PushJobEvent(jobEvent)
		}

	}

	//2.从该revision向后监听变化事件
	go func(){ //监听协程
		watchStartRevision = getResp.Header.Revision+1 //从GET时刻的后序revision监听
		//监听/cron/jobs/目录的后序变化
		watchChan = jobMgr.watcher.Watch(context.TODO(),common.JOB_SAVE_DIR,clientv3.WithRev(watchStartRevision),clientv3.WithPrefix())
		//处理监听事件
		for watchResp = range watchChan{
			for _,watchEvent = range watchResp.Events{
				switch watchEvent.Type{
				case mvccpb.PUT://任务保存事件
					if job,err = common.UnpackageJob(watchEvent.Kv.Value);err != nil{
						continue
					}
					//构造一个Event事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)

				case mvccpb.DELETE://任务被删除了
					//观察到Delete /cron/jobs/job10，我们实际要做的是将job10提取出来，去删除它
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))

					job = &common.Job{Name:jobName,}

					//构造一个删除Event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)

					
				}

				//TODO:推给scheduler
				G_scheduler.PushJobEvent(jobEvent)
			
			}
		}
	}()

	return 
}


 //InitJobMgr 初始化管理器--连接ETCD
 func InitJobMgr() (err error){
	 
	var (
		config clientv3.Config
		client *clientv3.Client
		kv   clientv3.KV
		lease clientv3.Lease
		watcher clientv3.Watcher
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


	//得到KV,lease,watcher的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	//赋值单例
	G_jobMgr = &JobMgr{
		client:client,
		kv:kv,
		lease:lease,
		watcher:watcher,
	}


	//启动任务监听
	G_jobMgr.watchJobs()
	return 
 }

