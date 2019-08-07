/*
 * @Descripttion: 配置文件解析类
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-05 21:28:14
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-07 19:38:21
 */

 package master

import (
	"encoding/json"
	"io/ioutil"
)
 
 //Config 程序配置
 type Config struct{
	APIPort 				int			`json:"apiPort"`
	APIReadTimeout 			int			`json:"apiReadTimeout"`
	APIWriteTimeout 		int			`json:"apiWriteTimeout"`
	EtcdEndpoints 			[]string 	`json:"etcdEndpoints"`
	EtcdDialTimeout 		int 		`json:"etcdDialTimeout"`
	WebRoot					string		`json:"webroot"`
	MongodbURI 				string		`json:"mongodbUri"`
	MongodbConnectTimeout 	int 		`json:"mongodbConnectTimeout"`
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