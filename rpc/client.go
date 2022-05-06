package rpc

import (
	"flag"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
)

var (
	// ClientMap RPC服务所对应的客户端
	ClientMap = map[string]*rpc.Client{}
	// Servers RPC服务
	Servers = Server{Adders: []string{}}
)

// 命令行参数RPCAdders=[localhost:1001 localhost:1002]设置RPC服务
func init() {
	flag.Var(&Servers, "RPCAdders", "RPC服务")
}

// StartRpcServer 创建RPC服务客户端
func StartRpcServer() error {
	for _, cluster := range Servers.Adders {
		client, err := jsonrpc.Dial("tcp", cluster)
		if err != nil {
			return err
		}
		log.Printf("[StartRpcServer]RPC client %s inited\n", cluster)
		ClientMap[cluster] = client
	}
	return nil
}

// ShutdownRpcServer 关闭所有RPC客户端
func ShutdownRpcServer() {
	for _, client := range ClientMap {
		err := client.Close()
		if err != nil && err != rpc.ErrShutdown {
			log.Printf("[ShutdownRpcServer]Close RPC client return %+v", err)
		}
	}
}
