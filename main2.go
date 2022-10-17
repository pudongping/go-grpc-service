package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	pb "github.com/pudongping/go-grpc-service/proto"
	"github.com/pudongping/go-grpc-service/server"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var grpcOrHttpPort string

func init() {
	flag.StringVar(&grpcOrHttpPort, "grpc_or_http_port", "8003", "一个端口上兼容多种协议，一个连接可以是 gRPC 或 HTTP 但不能同时是两者")
	flag.Parse()
}

func RunTCPServer2(port string) (net.Listener, error) {
	return net.Listen("tcp", ":"+port)
}

func RunGrpcServer2() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())

	// 注册反射服务，方便让 grpcurl 或者 grpcui 用作调试
	// reflection 包是 gRPC 官方所提供的反射服务
	reflection.Register(s)

	return s
}

func RunHttpServer2(port string) *http.Server {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})

	return &http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}
}

func main() {
	l, err := RunTCPServer2(grpcOrHttpPort)
	if err != nil {
		log.Fatalf("Run TCP Server err: %v", err)
	}

	m := cmux.New(l)
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := m.Match(cmux.HTTP1Fast())

	grpcS := RunGrpcServer2()
	httpS := RunHttpServer2(grpcOrHttpPort)
	go grpcS.Serve(grpcL)
	go httpS.Serve(httpL)

	log.Printf("grpc or http server is started at: 127.0.0.1:%s \n", grpcOrHttpPort)

	err = m.Serve()
	if err != nil {
		log.Fatalf("Run Serve err: %v", err)
	}
}

// 测试
// go run main2.go
// grpcurl -plaintext -d '{"name":"go"}' localhost:8003 proto.TagService/GetTagList
// curl http://127.0.0.1:8003/ping
