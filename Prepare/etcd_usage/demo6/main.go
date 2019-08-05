package main

import (
	"fmt"
	"time"
	"github.com/coreos/etcd/clientv3"
	"context"
)

func main(){

	var (
		config clientv3.Config
		client *clientv3.Client
		lease clientv3.Lease
		err error	
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId clientv3.LeaseID
		putResp *clientv3.PutResponse
		getResp *clientv3.GetResponse
		keepResp *clientv3.LeaseKeepAliveResponse
		keepRespChan <-chan *clientv3.LeaseKeepAliveResponse
		kv  clientv3.KV
	)

	config = clientv3.Config{
		Endpoints:[]string{"192.168.80.129:2379"},
		DialTimeout:5 * time.Second,
	}

	

	//建立一个客户端
	if client,err = clientv3.New(config);err != nil{
		fmt.Println(err)
		return 
	}

	defer client.Close()


	//申请一个租约lease
	lease = clientv3.NewLease(client)

	//申请一个10s的租约
	if leaseGrantResp,err = lease.Grant(context.TODO(), 10);err != nil{
		fmt.Println(err)
		return
	}

	
	//拿到租约的ID
	leaseId = leaseGrantResp.ID
	
	//5秒后会取消自动续租
	if keepRespChan,err = lease.KeepAlive(context.TODO(), leaseId);err != nil{
		fmt.Println(err)
		return 	
	}
	
	//处理续租应答的协程
	go func(){
		for{
			select{
			case keepResp = <- keepRespChan:
				if keepRespChan == nil{
					fmt.Println("租约已经失效（异常原因）")
					goto END
				}else{	//每秒会续租一次
					fmt.Println("收到自动续租应答：",keepResp.ID)
				}
			}
		}
		END:
	}()

	//获得kv对象
	kv = clientv3.NewKV(client)

	//Put一个KV，让它与租约关联起来，从而实现10s后自动过期
	if putResp,err = kv.Put(context.TODO(), "/cron/lock/job1","",clientv3.WithLease(leaseId));err != nil{
		fmt.Println(err)
	}

	fmt.Println("写入成功",putResp.Header.Revision)

	//定时看一下key过期了没有
	for{
		 if getResp,err = kv.Get(context.TODO(), "/cron/lock/job1");err != nil{
			 fmt.Println(err)
			 return
		}

		if getResp.Count == 0{
			fmt.Println("kv过期了")
			break
		}

		fmt.Println("还没过期",getResp.Kvs)
		time.Sleep(2 * time.Second)
	}


}