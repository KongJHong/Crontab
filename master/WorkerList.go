package master

import (
	"github.com/coreos/etcd/mvcc/mvccpb"
	"context"
	"Crontab/common"
	"github.com/coreos/etcd/clientv3"
	"time"
)

// WorkerMgr worker管理类
type WorkerMgr struct{
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease
}


var (
	G_workerMgr *WorkerMgr
)


//ListWorkers 获取在worker列表
func (workerMgr *WorkerMgr)ListWorkers()(workerArr []string,err error){

	var (
		getResp *clientv3.GetResponse
		workerIP string
		kv *mvccpb.KeyValue
	)
	
	//初始化数组
	workerArr = make([]string,0)

	//获取目录下的所有kv
	if getResp,err = workerMgr.kv.Get(context.TODO(), common.JOB_WORKER_DIR, clientv3.WithPrefix());err != nil{
		return 
	}

	//解析每个节点的IP
	for _,kv = range getResp.Kvs{
		//kv.key: /cron/workers/192.168.1.2
		workerIP = common.ExtractWorkerIP(string(kv.Key))
		workerArr = append(workerArr,workerIP)
	}

	return
}



//InitWorkerMgr worker发现初始化
func InitWorkerMgr()(err error){

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

	//得到KV,lease,watcher的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	G_workerMgr = &WorkerMgr{
		client:client,
		kv:kv,
		lease:lease,
	}

	return
}