package main

import (
	"github.com/mongodb/mongo-go-driver/bson/objectid"
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

func main(){

	var (
		client *mongo.Client
		err error
		database *mongo.Database
		collection *mongo.Collection
		record *LogRecord
		result *mongo.InsertOneResult
		docId objectid.ObjectID
	)
	
	//1.建立连接
	if client,err=mongo.Connect(context.TODO(),"mongodb://120.79.57.186:27017",clientopt.ConnectTimeout(5 * time.Second));err != nil{
		fmt.Println(err)		
	}

	//2.选择数据库my_db
	database = client.Database("cron")

	//3.选择表my_collection
	collection = database.Collection("log")

	//4.插入记录(bson)
	record = &LogRecord{
		JobName:"job10",
		Command:"echo hello",
		Err:"",
		Content:"hello",
		TimePoint:TimePoint{StartTime:time.Now().Unix(),EndTime:time.Now().Unix()+10,},
	}

	if result,err = collection.InsertOne(context.TODO(), record);err != nil{
		fmt.Println(err)
		return
	}

	//_id:默认生成一个全局唯一ID，ObjectID:12字节的二进制
	docId = result.InsertedID.(objectid.ObjectID)	//必须反射出来
	fmt.Println("自增ID: ",docId.Hex(),"-",docId.String())
}