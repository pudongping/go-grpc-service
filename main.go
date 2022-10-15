package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/pudongping/go-grpc-service/proto"
	"github.com/pudongping/go-grpc-service/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	s := grpc.NewServer()

	pb.RegisterTagServiceServer(s, server.NewTagServer())

	port := 5200
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	log.Printf("grpc server is started at: 127.0.0.1:%d", port)
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}

	// 注册反射服务，方便让 grpcurl 或者 grpcui 用作调试
	// reflection 包是 gRPC 官方所提供的反射服务
	reflection.Register(s)

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("server.Serve err: %v", err)
	}
}
