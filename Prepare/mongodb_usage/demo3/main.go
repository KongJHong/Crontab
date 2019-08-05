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
		logArr []interface{} 	//C语言里的void* ,记录了type
		result *mongo.InsertManyResult
		insertId interface{}		//_id:
		docId objectid.ObjectID
	)
	
	//1.建立连接
	if client,err=mongo.Connect(context.TODO(),"mongodb://192.168.80.129:27017",clientopt.ConnectTimeout(5 * time.Second));err != nil{
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

	//5.批量插入多条document
	logArr = []interface{}{record,record,record}
	
	//发起插入
	if result,err = collection.InsertMany(context.TODO(), logArr);err != nil{
		fmt.Println(err)
		return
	}

	//twitter很早的时候的开源的，用来算twitter的ID
	//insertID由snowflake算法计算生成
	for _,insertId = range result.InsertedIDs{
		//拿着interface{}反射成objectID
		docId = insertId.(objectid.ObjectID)
		fmt.Println("自增ID:",docId.Hex())
	}
}