/*
 * @Descripttion: 与MongoDB进行连接
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-07 15:28:42
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-07 17:02:52
 */

package worker

import (
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"context"
	"Crontab/common"
	"github.com/mongodb/mongo-go-driver/mongo"
	"time"
)

//LogSink mongodb存储日志
type LogSink struct{
	client *mongo.Client
	logCollection *mongo.Collection
	logChan chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	//单例
	G_logSink *LogSink
)

//saveLogs 批量写入日志
func (logSink *LogSink)saveLogs(batch *common.LogBatch){
	logSink.logCollection.InsertMany(context.TODO(), batch.Logs)
}


//writeLoop 日志存储协程
func (logSink *LogSink)writeLoop(){
	var (
		log *common.JobLog
		logBatch *common.LogBatch//当前批次
		commitTimer *time.Timer
		timeoutBatch *common.LogBatch	//超时批次
	)

	for{
		select{
		case log = <-logSink.logChan:
			//把这个log写到mongodb中
			//logSink.logCollection.InsertOne操作就好了,但是这样做太慢了
			//每次插入需要等待mongodb的一次请求往返，耗时可能因为网络慢话费时间比较长的时间
			//所以应该批处理
			if logBatch == nil{
				logBatch = &common.LogBatch{}
				//让这次超时自动提交（给1秒的时间）
				commitTimer = time.AfterFunc(
					time.Duration(G_config.JobLogCommitTimeout) * time.Millisecond,
					func(batch *common.LogBatch) func() {
						//这个函数会帮你在另外一个协程执行，那么另外一个协程也提交batch，这个writeLoop也提交batch，这种时候应该串行化处理
						//发出超时通知，不要直接提交batch
						//logSink.autoCommitChan<-logBatch 这里的logBatch是外部的，随时可能被改掉，所以要另寻办法，可以从外部传进来！！！
						return func(){
							logSink.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}
			//把新的日志插入到批次中去
			logBatch.Logs = append(logBatch.Logs,log)

			//如果批次满了，立即发送.这种情况应该有时间的约束，不能干等100条
			if len(logBatch.Logs) >= G_config.JobLogBatchSize{
				//发送日志
				logSink.saveLogs(logBatch)
				//清空logBatch
				logBatch = nil
				//取消定时器，防止1秒循环
				commitTimer.Stop()
			}
		case timeoutBatch = <- logSink.autoCommitChan://过期的批次
			//判断过期批次仍旧是当前批次
			if timeoutBatch != logBatch{ //!= 说明logBatch已经被上面的逻辑清空了，这里就不应该继续处理了
				continue	//跳过已经被提交的批次
			}
			//把这个批次写入到mongo中去
			logSink.saveLogs(timeoutBatch)
			//清空logBatch
			logBatch = nil
		}
	}	
}

//InitLogSink 初始化sink
func InitLogSink()(err error){
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
	G_logSink = &LogSink{
		client:client,
		logCollection:client.Database("cron").Collection("log"),
		logChan:make(chan *common.JobLog,G_config.MemoryCache),	
		autoCommitChan:make(chan *common.LogBatch,1000),
	}

	//启动一个Mongdb处理协程
	go G_logSink.writeLoop()
	return 
}

//Append 发送日志API
func (logSink *LogSink)Append(jobLog *common.JobLog){
	select{
	case logSink.logChan <- jobLog:
	default:
		//队列满了就丢弃
	}
	
}