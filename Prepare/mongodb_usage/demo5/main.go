package main

import (
	"fmt"
	"time"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"context"
	"github.com/mongodb/mongo-go-driver/mongo"
)

//TimePoint 记录任务的执行时间点
type TimePoint struct{
	StartTime int64	`bson:"startTime"`
	EndTime  int64	`bson:"endTime"`
}


//LogRecord 一条日志的结构体
type LogRecord struct{
	JobName string `bson:"jobName"` 			//任务名
	Command	string `bson:"command"`			//shell命令
	Err		string `bson:"err"`				//错误
	Content string `bson:"content"`			//脚本输出
	TimePoint TimePoint `bson:"timepoint"` 	//执行时间点
}

//FindByJobName jobName过滤条件
type FindByJobName struct{
	JobName string `bson:"jobName"`		//JobName复制为job10
}

//startTime 小于某时间
//{"$lt":timestamp}
type TimeBeforeCond struct{
	Before int64 `bson:"$lt"`
}

//DeleteCond {"timePoint.startTime":{"$lt":timestamp}}
type DeleteCond struct{
	beforeCond TimeBeforeCond `bson:"timePoint.startTime"`
}

func main(){

	//mongodb读取回来的是bson,需要反序列化为LogRecord对象

	var (
		client *mongo.Client
		err error
		database *mongo.Database
		collection *mongo.Collection
		delCond *DeleteCond
		delResult *mongo.DeleteResult
	)
	
	//1.建立连接
	if client,err=mongo.Connect(context.TODO(),"mongodb://192.168.80.129:27017",clientopt.ConnectTimeout(5 * time.Second));err != nil{
		fmt.Println(err)		
	}

	//2.选择数据库my_db
	database = client.Database("cron")

	//3.选择表my_collection
	collection = database.Collection("log")

	//4.要删除开始时间早于当前时间的所有日志
	//delete({"timePoint.startTime":{"$lt":当前时间}})
	delCond = &DeleteCond{beforeCond:TimeBeforeCond{Before:time.Now().Unix()}}

	//执行删除
	if delResult,err = collection.DeleteMany(context.TODO(), delCond);err != nil{
		fmt.Println(err)
		return
	}

	fmt.Println("删除了",delResult.DeletedCount,"行")

}
