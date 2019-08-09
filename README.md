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
- nginx负载均衡

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/{517CB287-70CA-74B3-664A-9B2631C8E0DD}.jpg)

### 项目版本

#### mongodb4.0.0

#### etcd-v3.3.8

1. 解压etcd
2. cd进入，输入 nohup ./etcd --listen-client-urls 'http://0.0.0.0:2379' --advertise-client-urls 'http://0.0.0.0:2379' &

### 项目结构

```
——————Crontab
 |
 |---master：master框架，主要管理路由等前端逻辑
 |     |---main:程序启动文件夹
 |     |    |---master.go:程序启动main主文件
 |     |    |---master.json:配置文件
 |     |
 |     |---webroot:静态页面文件夹
 |     |    |---index.html:前端主界面，调用master的API
 |     |
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







