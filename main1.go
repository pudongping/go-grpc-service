package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "github.com/pudongping/go-grpc-service/proto"
	"github.com/pudongping/go-grpc-service/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var grpcPort string
var httpPort string

func init() {
	flag.StringVar(&grpcPort, "grpc_port", "8001", "gRPC 启动端口号")
	flag.StringVar(&httpPort, "http_port", "8002", "HTTP 启动端口号")
	flag.Parse()
}

func main() {
	errs := make(chan error)

	go func() {
		err := RunHttpServer1(httpPort)
		if err != nil {
			errs <- err
		}
	}()

	go func() {
		err := RunGrpcServer1(grpcPort)
		if err != nil {
			errs <- err
		}
	}()

	select {
	case err := <-errs:
		log.Fatalf("Run Server err: %v", err)
	}
}

func RunHttpServer1(port string) error {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})

	log.Printf("http server is started at: 127.0.0.1:%s \n", port)
	return http.ListenAndServe(":"+port, serveMux)
}

func RunGrpcServer1(port string) error {
	s := grpc.NewServer()

	pb.RegisterTagServiceServer(s, server.NewTagServer())

	// 注册反射服务，方便让 grpcurl 或者 grpcui 用作调试
	// reflection 包是 gRPC 官方所提供的反射服务
	reflection.Register(s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	log.Printf("grpc server is started at: 127.0.0.1:%s \n", port)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
		return err
	}

	return s.Serve(lis)
}
