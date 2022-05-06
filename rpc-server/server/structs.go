package server

import (
	bytes2 "bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// RpcFun RPC对象
type RpcFun string

// Echo RPC对象导出方法 负责对RPC请求进行响应 类似于HTTP中的Handler
func (*RpcFun) Echo(args *Args, reply *Reply) error {
	/*
		可以从args中获取各种参数 进行业务逻辑
	*/

	// 需要设置在HTTP请求头的key-value
	reply.HeaderKeyValues = map[string]string{"RPCServer": Addr}
	// 将请求body数据解码到body中 (可以有多种解码方式 暂时默认该接口使用Json解码编码)
	body := make(map[string]interface{})
	err := decodeJSON(bytes2.NewReader(args.Body), &body)
	if err != nil {
		// 设置响应结构体中的HTTP响应码
		reply.HttpStatusCode = http.StatusBadRequest
		return err
	}
	body["RPCServer"] = Addr
	// 将响应数据编码 (可以有多种编码方式 暂时默认该接口使用Json解码编码)
	bytes, err := encodeJson(body)
	if err != nil {
		// 设置响应结构体中的HTTP响应码
		reply.HttpStatusCode = http.StatusBadRequest
		return err
	}
	reply.Body = bytes
	// 设置响应结构体中的HTTP响应码
	reply.HttpStatusCode = http.StatusOK
	return nil
}

// Args RPC请求结构体
type Args struct {
	Method           string
	URL              *url.URL
	Proto            string // "HTTP/1.0"
	Header           http.Header
	Body             []byte
	ContentLength    int64
	TransferEncoding []string
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	TLS              *tls.ConnectionState
	Response         *http.Response
}

// Reply RPC响应结构体
type Reply struct {
	HttpStatusCode  int
	HeaderKeyValues map[string]string
	Body            []byte
}

// 将r中数据解码到obj
func decodeJSON(r io.Reader, obj interface{}) error {
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return nil
}

// 将obj数据编码为json返回
func encodeJson(obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}
