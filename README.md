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
 |
 |
 |
 |
 |---worker
 |
 |
 |
 |---common：共享的类或者结构
 |     |---Protocol.go:保存一些交互协议类
 |     |---Constants.go:系统公用常量

```

#### Master

- 搭建go项目框架，配置文件，命令行参数，线程配置
- 给web后台提供http API,用于管理job
- 写一个web后台的前端页面，bootstrap+jquery，前后端分离开发

#### Worker

- 从etcd中把job同步到内存中
- 实现调度模块，基于cron表达式调度N个job
- 实现执行模块，并发的执行多个job
- 对job的分布式锁，防止集群并发
- 把执行日志保存到`mongodb`



**Golang执行原理**

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/{6AA88781-375D-16DA-E4B9-99EF7CBC2FAD}.jpg)

### 传统crontab痛点

- 机器故障，任务停止调度，甚至`crontab`配置都找不回来
- 任务数量多，单机的硬件资源耗尽，需要人工迁移到其他机器
- 需要人工去机器上配置`cron`，任务执行状态下不方便查看

### 分布式架构 - 核心要素

- 调度器：需要高可用，确保不会因为单点故障停止调度（调度的高可用）
- 执行器：需要扩展性，提供大量任务的并行处理能力（执行的可扩展）

### 常见开源调度架构

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805191629.png)

#### 伪分布式设计

- 分布式网络环境不可靠，`RPC`异常属于常态
- Master下发任务`RPC`，导致Master与Worker状态不一致
- Worker上报任务`RPC`异常，导致Master状态信息落后

#### 异常case举例

- 状态不一致：Master下发任务给node1异常，实际node1收到并开始执行
- 并发执行：Master重试下发任务给node2，结果node1与node2同时执行一个任务
- 状态丢失：Master更新zookeeper中任务状态异常，此时Master主机切换Standby，任务仍旧处于状态

#### CAP理论（常用于分布式存储）

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805192614.png)

CAP理论中P几乎是必须实现的，所以一般都在CA中取舍

#### BASE理论（常用于应用架构）

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805192958.png)

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805193127.png)

### MASTER-WORKER整体架构

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805195523.png)

#### 架构思路

- 利用etcd同步全量任务列表到所有worker节点
- 每个worker独立调度全量任务，无需与master产生直接RPC
- 各个worker利用分布式锁抢占，解决并发调度解决相同任务的问题

### Master功能点和实现思路

#### Master功能

- 任务管理HTTP接口：新建、修改、查看、删除任务
- 任务日志HTTP接口：查看任务执行历史日志
- 任务控制HTTP接口：提供强制结束任务的接口
- 实现web管理界面：基于jquery+bootstrap的Web控制台，前后端分离

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805200138.png)

**Etcd结构**

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805200203.png)

**任务管理**

- 保存到etcd的任务，会被实时同步到所有的worker

**任务日志**

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805200344.png)

- master请求MongoDB,按任务名查看最近的执行日志

**任务控制**

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805200440.png)

想到与到worker集群的通知

- worker监听`/cron/killer/`目录下put修改操作
- master将要结束的任务名put在`/cron/killer/`目录下，触发worker立即技术shell任务

### Worker功能点与实现思路

**Worker功能**

- 任务同步：监听`/cron/jobs/`目录变化
- 任务调度：基于`cron`表达式计算，触发过期任务
- 任务执行：协程池并发执行多任务，基于etcd分布式锁抢占
- 日志捕获：捕获任务执行输出，保存到MongoDB

**Worker内部架构**

![](https://kongjhong-image.oss-cn-beijing.aliyuncs.com/img/20190805201941.png)

**监听协程**

- 利用watch API,监听`/cron/jobs/`和`/cron/killer/`目录的变化
- 将变化事件通过channel推送给调度协程，更新内存中的任务信息

**调度协程**

- 监听任务变成event,更新内存中维护的任务列表
- 检查任务cron表达式，扫描到期的任务，交给执行协程运行
- 监听任务控制event，强制中断正在执行中的子进程
- 监听任务执行result，更新内存中的任务状态，投递执行日志

**执行协程**

- 在etcd中抢占分布式乐观锁:`/cron/lock/任务名`
- 抢占成功则通过Command类执行shell任务
- 捕获Command输出并等待子进程结束，将执行结果投递给调度协程

**日志协程**

- 监听调度发来的执行日志，放入一个batch中
- 对新batch启动定时器，超时未满自动提交
- 若batch被放满，那么立即提交，并取消自动提交定时器







