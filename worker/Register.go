package worker

import (
	"context"
	"Crontab/common"
	"net"
	"github.com/coreos/etcd/clientv3"
	"time"
)


//Register 注册节点到ETCD: /cron/workers/IP地址
type Register struct{
	client *clientv3.Client
	kv clientv3.KV
	lease clientv3.Lease

	localIP string	//本机Ip
}

var (
	G_register *Register
)


//获取本机网卡IP
func getLocalIP()(ipv4 string,err error){
	
	var (
		addrs []net.Addr
		addr net.Addr
		ipNet *net.IPNet	//IP地址
		isIPNet bool
	)

	//遍历所有的网卡
	if addrs,err = net.InterfaceAddrs();err != nil{
		return 
	}

	//取第一个非localhost的网卡
	for _,addr = range addrs{
		//ipv4,ipv6
		//这个网络地址是IP地址
		if ipNet,isIPNet = addr.(*net.IPNet);isIPNet && !ipNet.IP.IsLoopback(){//是ip地址，且不是环回网卡
			//跳过IPV6
			if ipNet.IP.To4() != nil{
				ipv4 = ipNet.IP.String()
				return 
			}
		}
	}
	
	err = common.ERR_NO_LOCAL_IP_FOUND
	return 
}

//keepOnline 自动注册到/cron/workers/IP,并自动续租
func (register *Register)keepOnline(){
	var (
		regKey string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err error
		keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp *clientv3.LeaseKeepAliveResponse
		cancelCtx context.Context
		cancelFunc context.CancelFunc
	)
	for{
		//注册路径
		regKey = common.JOB_WORKER_DIR + register.localIP

		cancelFunc = nil

		//创建租约
		if leaseGrantResp,err = register.lease.Grant(context.TODO(), 10);err != nil{
			//异常就应该重试
			goto RETRY	
		}

		//自动续租
		if keepAliveChan,err = register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID);err != nil{
			goto RETRY
		}

		//只有注册失败时才取消租约，因为租约的达成是通过与ETCD建立才生效的
		cancelCtx,cancelFunc = context.WithCancel(context.TODO())

		//建立注册到ETCD
		if _,err = register.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID));err != nil{
			goto RETRY	
		}

		//处理续租应答
		for {
			select{
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil{	//续租失败
					goto RETRY
				}
			}
		}

		RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil{
			cancelFunc()
		}
	}

}


//InitRegister 注册初始化
func InitRegister()(err error){

	var (
		config clientv3.Config
		client *clientv3.Client
		kv   clientv3.KV
		lease clientv3.Lease
		localIP string
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints :G_config.EtcdEndpoints,	//集群地址
		DialTimeout : time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,	//连接超时
	}

	//建立连接
	if client,err = clientv3.New(config);err != nil{
		return
	}

	//本机IP
	if localIP,err = getLocalIP();err != nil{
		return 
	}

	//得到KV,lease,watcher的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)


	G_register = &Register{
		client:client,
		kv:kv,
		lease:lease,
		localIP:localIP,
	}

	//服务注册
	go G_register.keepOnline()

	return 
}