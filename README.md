<!--
 * @Descripttion: 
 * @version: 
 * @Author: KongJHong
 * @Date: 2019-08-04 09:39:21
 * @LastEditors: KongJHong
 * @LastEditTime: 2019-08-09 22:35:03
 -->
## Crontab 分布式任务调度

### 主要概念

- master-worker分布式架构
- 服务注册于发现
- 任务分发
- 分布式锁
- CAP理论
- etcd协调服务
- Raft协议
- 事件广播
- 多任务调度
- 异步日志
- 并发设计
- mongodb分布式存储
- systemclt服务管理

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/{517CB287-70CA-74B3-664A-9B2631C8E0DD}.jpg)

### 项目版本

#### mongodb4.0.0
1. 解压mongodb
2. cd mongodb
3. mkdir data
4. nohup bin/mongod --dbpath=./data --bind_ip=0.0.0.0 & 后台运行
5. bin/mongo 进入数据库

#### etcd-v3.3.8

1. 解压etcd
2. cd进入，输入 nohup ./etcd --listen-client-urls 'http://0.0.0.0:2379' --advertise-client-urls 'http://0.0.0.0:2379' &

### 项目结构

```
——————Crontab
 |
 |---master：master框架，主要管理路由等前端逻辑
 |     |---main:程序启动文件夹
       |    |---webroot:静态页面文件夹
 |     |    |     |---index.html:前端主界面，调用master的API
 |     |    |---master.go:程序启动main主文件
 |     |    |---master.json:配置文件
 |     |
 |     |---ApiServer.go:HTTP路由管理，前端到后台任务的CRUD
 |     |---Config.go:程序配置类，读取main/master.json中的配置
 |     |---JobMgr.go:任务管理类，实际管理任务的增删改查（与ETCD交互）
 |     |---LogSink.go:日志持久化保存，连接Mongodb
 |     |---LogMgr.go:日志查看
 |
 |---worker
 |     |---main:worker程序启动文件夹
 |     |    |---worker.go: main文件
 |     |    |---worker.json:worker配置文件
 |     |
 |     |---Config.go: worker配置类，读取main/worker.json的配置
 |     |---JobMgr.go:任务管理类，设置ETCD监听任务，把变化任务推到Scheduler中
 |     |---Scheduler.go:任务调度功能类，包含定时器，任务执行
 |     |---Executor.go:执行器，执行任务
 |     |---JobLock.go:分布式锁，抢占分布式锁
 |     |---Register.go:worker服务注册到ETCD
 |
 |---common：共享的类或者结构
 |     |---Protocol.go:保存一些交互协议类
 |     |---Constants.go:系统公用常量
 |     |---Errors.go:可复用的错误信息

```



### 项目启动

确定已经配置好Golang开发环境，本程序可在Linux和Windows环境运行，ETCD和MongoDB需要架设在Linux主机环境上

分别修改`/Crontab/master/main/master.json 以及 /Crontab/worker/main/worker.json`文件的IP地址

下面以Linux环境进行演示

```cpp
git clone https://github.com/KongJHong/Crontab.git

cd Crontab

export $GOPATH=`pwd` 

cd /Crontab/master/main/

修改/Crontab/master/main/master.json 为对应IP地址

go build

./main -config ./master.json 运行起来

worker同理
```

