/*
 * @Descripttion: master日志管理类
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-07 19:35:05
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-07 20:02:31
 */

package master

import (
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
	"Crontab/common"
	"time"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"context"
)

//LogMgr mongodb日志管理
type LogMgr struct{
	client *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)


//InitLogMgr 初始化sink
func InitLogMgr()(err error){
	var (
		client *mongo.Client
	)

	//建立mongodb连接
	if client,err = mongo.Connect(context.TODO(), 
						G_config.MongodbURI, 
						clientopt.ConnectTimeout(time.Duration(G_config.MongodbConnectTimeout) * time.Millisecond));err != nil{
		return 
	}


	//选择db和connection
	G_logMgr = &LogMgr{
		client:client,
		logCollection:client.Database("cron").Collection("log"),
	}

	//启动一个Mongdb处理协程
	
	return 
}

//ListLog 查看任务日志
func (logMgr *LogMgr)ListLog(name string,skip int,limit int) (logArr []*common.JobLog,err  error){
	
	var (
		filter *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor mongo.Cursor	
		jobLog *common.JobLog
	)

	//len(logArr),提高代码健壮性
	logArr = make([]*common.JobLog,0)

	//过滤条件
	filter = &common.JobLogFilter{JobName:name}

	//按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{SortOrder:-1}

	//发起查询
	if cursor,err = logMgr.logCollection.Find(context.TODO(), filter, findopt.Sort(logSort),findopt.Skip(int64(skip)),findopt.Limit(int64(limit)));err != nil{
		return 
	}

	//延迟释放游标
	defer cursor.Close(context.TODO())

	//遍历游标
	for cursor.Next(context.TODO()){
		jobLog = &common.JobLog{}

		//反序列化BSON
		if err = cursor.Decode(jobLog);err != nil{
			continue //有日志不合法
		}

		logArr =append(logArr,jobLog)
		
	}
	
	

	return 
}