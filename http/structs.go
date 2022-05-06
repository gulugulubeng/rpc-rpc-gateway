package http

import (
	"crypto/tls"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// DefaultTimeout 默认超时时长
const DefaultTimeout = time.Second * 10

// Args RPC请求结构体
type Args struct {
	Method           string      // 请求方法
	URL              *url.URL    // HTTP URL
	Proto            string      // "HTTP/1.0"
	Header           http.Header // HTTP请求头
	Body             []byte      // HTTP请求体
	ContentLength    int64       // 请求体长度
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
	HttpStatusCode  int               // HTTP响应状态码
	HeaderKeyValues map[string]string // HTTP响应头key-value
	Body            []byte            // HTTP响应体
}

// PoolArg 协程池任务
type PoolArg struct {
	Rw       http.ResponseWriter
	Req      *http.Request
	Finished chan bool
}
