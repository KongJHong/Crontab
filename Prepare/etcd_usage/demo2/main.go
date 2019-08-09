package main

import (
	"context"
	"fmt"
	"time"
	"github.com/coreos/etcd/clientv3"
)


func main(){

	var(
		config clientv3.Config
		client *clientv3.Client
		err error
		kv clientv3.KV
		//ctx context.Context
		putResp *clientv3.PutResponse
	)

	config = clientv3.Config{
		Endpoints:[]string{"120.79.57.186:2379"},
		DialTimeout:5 * time.Second,
	}

	//建立一个客户端
	if client,err = clientv3.New(config);err != nil{
		fmt.Println(err)
		return 
	}

	//用于读写etcd的键值对
	kv = clientv3.NewKV(client)


	fmt.Println(".....")  //kv.Put(...) 一直阻塞不返回结果
	if putResp,err = kv.Put(context.TODO(), "/cron/jobs/job1", "bye",clientv3.WithPrevKV());err !=nil{
		fmt.Println(err)
	}else{
		fmt.Println("Revision",putResp.Header.Revision)
		if putResp.PrevKv != nil{ //打印Hello
			fmt.Println("PrevValue:",string(putResp.PrevKv.Value))
		}
	}
}