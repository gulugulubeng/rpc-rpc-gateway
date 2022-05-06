package rpc

import "strings"

// Server RPC服务集合
type Server struct {
	Adders []string
}

func (r *Server) String() string {
	res := strings.Builder{}
	for i, adder := range r.Adders {
		res.WriteString(adder)
		if i != len(r.Adders)-1 {
			res.WriteString(",")
		}
	}
	return res.String()
}

func (r *Server) Set(s string) error {
	r.Adders = strings.Split(s, ",")
	return nil
}
