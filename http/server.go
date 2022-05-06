package http

import (
	"context"
	"flag"
	"log"
	"net/http"
	"rpc-gateway/pool"
)

var (
	server   *http.Server
	httpAddr string
	httpPool *pool.Pool
)

// 命令行参数HTTPAddr=localhost:80设置HTTP监听端口
func init() {
	flag.StringVar(&httpAddr, "HTTPAddr", "localhost:80", "HTTP监听地址")
}

// StartHttpServer 开启HTTP服务
func StartHttpServer() error {
	var err error
	// 初始化协程池
	httpPool, err = pool.NewPool(func(req interface{}) {
		// 断言请求参数
		arg := req.(*PoolArg)
		// 处理HTTP请求
		HttpHandler(arg.Rw, arg.Req)
		// 处理完毕信号
		arg.Finished <- true
	})
	if err != nil {
		return err
	}

	// 初始化HTTP服务
	server = &http.Server{
		Addr: httpAddr,
	}

	// 注册处理路由
	http.HandleFunc("/", HttpHandler)         // 正常处理HTTP请求
	http.HandleFunc("/pool", HttpPoolHandler) // 使用协程池处理HTTP请求

	// 异步开启HTTP服务开始阻塞监听
	go func() {
		log.Printf("[StartHttpServer]HTTP server addr %s", httpAddr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("[StartHttpServer]HTTP server listen and serve return %+v", err)
		}
	}()
	return nil
}

// ShutdownHttpServer 关闭HTTP服务
func ShutdownHttpServer() {
	// 关闭协程池
	if httpPool != nil {
		err := httpPool.Release()
		if err != nil {
			log.Printf("[ShutdownHttpServer]Pool release return %+v", err)
		}
	}

	// 关闭HTTP服务
	if server != nil {
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Printf("[ShutdownHttpServer]Shutdown http server return %+v", err)
		}
	}
}
