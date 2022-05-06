package server

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

var (
	Addr     string
	listener net.Listener
)

// 命令行参数Addr=localhost:1001设置RPC服务地址
func init() {
	flag.StringVar(&Addr, "Addr", "localhost:1001", "RPC服务地址")
}

// StartRPCServer RPC服务阻塞监听
func StartRPCServer() error {
	// 注册RPC服务函数(可注册多个函数)
	err := rpc.Register(new(RpcFun))
	if err != nil {
		log.Printf("[StartRPCServer]RPC register fun return %+v", err)
		return err
	}
	// json rpc 服务端
	listener, err = net.Listen("tcp", Addr)
	if err != nil {
		log.Printf("[StartRPCServer]net create listener return %+v", err)
		return err
	}
	//服务端等待请求
	log.Printf("[StartRPCServer]RPC listen addr %s \n", Addr)
	// 异步阻塞监听
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("[StartRPCServer]RPC listener accept conn return %+v \n", err)
				break
			}
			log.Printf("[StartRPCServer]RPC accept conn from %s \n", conn.RemoteAddr())
			// 并发处理客户端请求
			go jsonrpc.ServeConn(conn)
		}
	}()
	return nil
}

// ShutDownRPCServer 关闭RPC服务
func ShutDownRPCServer() error {
	return listener.Close()
}
