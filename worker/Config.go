/*
 * @Descripttion: worker配置文件解析类
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 21:28:14
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-07 19:37:56
 */

 package worker

import (
	"encoding/json"
	"io/ioutil"
)
 
 //Config 程序配置
 type Config struct{
	EtcdEndpoints 			[]string 	`json:"etcdEndpoints"`
	EtcdDialTimeout 		int 		`json:"etcdDialTimeout"`
	ShellLocation			string  	`json:"shellLocation"`
	MongodbURI 				string		`json:"mongodbUri"`
	MongodbConnectTimeout 	int 		`json:"mongodbConnectTimeout"`
	MemoryCache 			int 		`json:"memoryCache"`
	JobLogBatchSize			int			`json:"jobLogBatchSize"`
	JobLogCommitTimeout		int 		`json:"jobLogCommitTimeout"`
	
 }


var (
	//单例
	G_config *Config
)



//InitConfig 加载配置
 func InitConfig(filename string)(err error){

	var (
		content []byte
		conf Config
	)
	//1.把配置文件读进来
	if content,err = ioutil.ReadFile(filename);err != nil{
		return 
	}


	//2.JSON反序列化
	if err = json.Unmarshal(content, &conf);err != nil{
		return 
	}

	//3.赋值单例
	G_config = &conf

	//fmt.Println(conf)
	return 
 }