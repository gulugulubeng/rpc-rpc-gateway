package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"rpc-gateway/rpc"
	"time"
)

// HttpPoolHandler 使用协程池处理HTTP
func HttpPoolHandler(rw http.ResponseWriter, req *http.Request) {
	// 需要协程池处理的任务
	arg := PoolArg{
		Rw:       rw,
		Req:      req,
		Finished: make(chan bool, 1),
	}
	if httpPool == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("协程池为空"))
		return
	}

	// 向协程池提交任务
	err := httpPool.Submit(&arg)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("协程池处理异常 " + err.Error()))
		return
	}

	// 阻塞等待协程池处理结果
	select {
	case <-arg.Finished: // 协程池处理完成
		rw = arg.Rw
	case <-time.After(DefaultTimeout): // 协程池处理超时
		rw.WriteHeader(http.StatusGatewayTimeout)
		rw.Write([]byte("协程池处理超时"))
	}
}

/**
协议转换HTTP请求转换为RPC请求

在HTTP请求头设置
	RPC服务地址：cluster
	RPC请求函数：rpcMethod

RPC请求设置
	请求参数为：*http.Request
	响应值接受：*http.ResponseWriter
*/

// HttpHandler 处理HTTP
func HttpHandler(rw http.ResponseWriter, req *http.Request) {
	// 从HTTP请求头获取RPC服务
	clu := req.Header.Get("cluster")
	if clu == "" || rpc.ClientMap[clu] == nil {
		rw.Write([]byte("目标RPC服务不存在"))
		return
	}
	// 从HTTP请求头获取rpc请求函数
	if req.Header.Get("rpcMethod") == "" {
		rw.Write([]byte("RPC函数无效"))
		return
	}

	// RPC请求
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte(fmt.Sprintf("读取Body错误: %+v", err)))
		return
	}
	// 初始化RPC请求参数
	args := &Args{
		Method:           req.Method,
		URL:              req.URL,
		Proto:            req.Proto,
		Header:           req.Header,
		Body:             bytes,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Host:             req.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
		TLS:              req.TLS,
		Response:         req.Response,
	}
	// 声明RPC响应结果
	var reply Reply
	// 异步发起RPC请求
	call := rpc.ClientMap[clu].Go(req.Header.Get("rpcMethod"), args, &reply, nil)
	// 阻塞等待RPC响应
	select {
	// 阻塞等待RPC请求完成
	case <-call.Done:
		// RPC返回异常
		err = call.Error
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(fmt.Sprintf("调用RPC错误: %+v", err)))
			return
		}
		// RPC返回正常
		rw.WriteHeader(reply.HttpStatusCode)      // 初始化响应状态码
		for k, v := range reply.HeaderKeyValues { // 设置HTTP响应头
			rw.Header().Set(k, v)
		}
		rw.Write(reply.Body) // 获取响应体
		// 阻塞等待超时
	case <-time.After(DefaultTimeout): // 默认超时时长
		rw.WriteHeader(http.StatusGatewayTimeout)
		rw.Write([]byte("RPC请求超时"))
	}
}
