package main

import (
	"github.com/coreos/etcd/mvcc/mvccpb"
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
		delResp *clientv3.DeleteResponse
	//	idx int
		kvpair *mvccpb.KeyValue
		
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

	//用于读写etcd的键值对
	kv = clientv3.NewKV(client)

	if delResp,err = kv.Delete(context.TODO(), "/cron/jobs/",clientv3.WithPrefix());err!=nil{
		fmt.Println(err)
		return 
	}

	//被删除之前的value是什么
	if len(delResp.PrevKvs) != 0{
		for _,kvpair = range delResp.PrevKvs{
			fmt.Println("删除了:",string(kvpair.Key),string(kvpair.Value))
		}
	}
}