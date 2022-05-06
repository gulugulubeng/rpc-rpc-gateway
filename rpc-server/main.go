package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"rpc-gateway/rpc-server/server"
	"syscall"
)

func main() {
	// 开启RPC服务
	err := server.StartRPCServer()
	if err != nil {
		fmt.Printf("Start RPC server return %+v \n", err)
	}

	// 阻塞主协程
	signals := make(chan os.Signal, 1)
	// 阻塞监听信号
	signal.Notify(signals, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT)
	fmt.Printf("System stop by signal:%+v \n", <-signals)

	// 关闭RPC服务
	err = server.ShutDownRPCServer()
	if err != nil {
		log.Printf("ShutDown RPC server return %+v \n", err)
	}
}

func init() {
	// 解析命令行参数
	flag.Parse()
}
