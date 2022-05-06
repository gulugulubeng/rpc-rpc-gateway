package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	myHttp "rpc-gateway/http"
	"rpc-gateway/rpc"
	"syscall"
)

func main() {
	// 初始化RPC服务客户端
	err := rpc.StartRpcServer()
	if err != nil {
		fmt.Printf("Start RPC server return %+v \n", err)
		return
	}
	// 开启HTTP服务
	myHttp.StartHttpServer()

	// 主进程阻塞
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)
	fmt.Printf("System stop by signal:%+v \n", <-signals)

	// 关闭HTTP服务及RPC客户端
	myHttp.ShutdownHttpServer()

	// 关闭RPC客户端服务
	rpc.ShutdownRpcServer()
}

func init() {
	// 解析命令行参数
	flag.Parse()
}
