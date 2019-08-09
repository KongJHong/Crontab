package cmdusage

import (
	"os/exec"
	"fmt"
	"context"
	"time"
)


func demo1and2(){
	var (
		cmd *exec.Cmd
		err error
		output []byte
	)
	//cmd = exec.Command("/bin/bash", "-c","echo1;echo2;")
	//生成Cmd
	cmd = exec.Command("C:\\cygwin64\\bin\\bash.exe", "-c","sleep 2;ls -l")

	//CombinedOutput() 执行了命令，捕获标准输出(pipe)
	if output,err = cmd.CombinedOutput();err != nil{
		fmt.Println(err)
		return
	}
	
	//err = cmd.Run()
	fmt.Println(string(output))
}

type result struct{
	output []byte
	err error
}

func main(){
	
	var (
		ctx context.Context
		cancelFunc context.CancelFunc 
		cmd *exec.Cmd
		resultChan chan *result
		res *result
	)

	//创建一个结果队列
	resultChan = make(chan *result,1000)

	//执行1个cmd，让它在一个协程中执行，让它执行2s，sleep 2;echo hello;

	//1秒的时候，我们杀死Cmd

	ctx,cancelFunc = context.WithCancel(context.TODO())

	go func(){
		var (
			output []byte
			err error
		)
		cmd = exec.CommandContext(ctx, "C:\\cygwin64\\bin\\bash.exe", "-c","sleep 2;echo hello")

		//执行任务，捕获输出
		output,err = cmd.CombinedOutput()

		//把任务输出结果，传给main协程
		resultChan<-&result{
			err : err,
			output : output,
		}
		close(resultChan)
	}()

	//继续往下走
	time.Sleep(1 * time.Second)

	//取消上下文
	cancelFunc()

	//在main协程里，等待子协程的退出，并打印任务执行结果
	res = <- resultChan
	fmt.Println(res.err,string(res.output))
}