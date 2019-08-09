package main

import (
	"context"
	"fmt"
	"time"
	"github.com/coreos/etcd/clientv3"
)


func main(){

	var (
		config clientv3.Config
		client *clientv3.Client
		err error
		kv clientv3.KV
		//getResp *clientv3.GetResponse
		putOp  clientv3.Op
		getOp  clientv3.Op
		opResp clientv3.OpResponse
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


	//创建Op
	putOp = clientv3.OpPut("/cron/jobs/job8", " 123135464")

	//执行op
	if opResp,err = kv.Do(context.TODO(), putOp);err != nil{
		fmt.Println(err)
		return
	}

	fmt.Println("写入revision:",opResp.Put().Header.Revision)

	//创建Op
	getOp = clientv3.OpGet("/cron/jobs/job8")

	//执行Op
	if opResp,err = kv.Do(context.TODO(), getOp);err != nil{
		fmt.Println(err)
		return
	}

	//打印
	fmt.Println("数据Revision:",opResp.Get().Kvs[0].ModRevision)   //create rev == mod rev
	fmt.Println("数据value:",string(opResp.Get().Kvs[0].Value))		

}