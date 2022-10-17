package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/pudongping/go-grpc-service/proto"
	"github.com/pudongping/go-grpc-service/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

var grpcAndHttpPort string

func init() {
	flag.StringVar(&grpcAndHttpPort, "grpc_and_http_port", "8004", "同端口下同 rpc 方法提供 gRPC 和 HTTP 双流量访问支持端口")
	flag.Parse()
}

func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	// gRPC 服务的非加密模式的设置：关注代码中的"h2c"标识，“h2c” 标识允许通过明文 TCP 运行 HTTP/2 的协议，
	// 此标识符用于 HTTP/1.1 升级标头字段以及标识 HTTP/2 over TCP，
	// 而官方标准库 golang.org/x/net/http2/h2c 实现了 HTTP/2 的未加密模式
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// gRPC 和 HTTP/1.1 的流量区分：
		// 对 ProtoMajor 进行判断，该字段代表客户端请求的版本号，客户端始终使用 HTTP/1.1 或 HTTP/2。
		// Header 头 Content-Type 的确定：grpc 的标志位 application/grpc 的确定。
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

func RunServer(port string) error {
	gatewayMux := runGrpcGatewayServer()

	httpMux := runHttpServer()

	grpcS := runGrpcServer()

	httpMux.Handle("/", gatewayMux)

	// ABC====>>>
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	s := &http.Server{
		Addr:    ":" + port,
		Handler: grpcHandlerFunc(grpcS, httpMux), // 请求的统一入口
	}
	return s.Serve(lis)
	// ABC====>>>

	// 和以上 ABC 区域代码一致
	// return http.ListenAndServe(":"+port, grpcHandlerFunc(grpcS, httpMux))
}

func runHttpServer() *http.ServeMux {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})

	return httpMux
}

func runGrpcServer() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterTagServiceServer(s, server.NewTagServer())

	// 注册反射服务，方便让 grpcurl 或者 grpcui 用作调试
	// reflection 包是 gRPC 官方所提供的反射服务
	reflection.Register(s)

	return s
}

func runGrpcGatewayServer() *runtime.ServeMux {
	endpoint := "0.0.0.0:" + grpcAndHttpPort
	gwmux := runtime.NewServeMux()
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := pb.RegisterTagServiceHandlerFromEndpoint(context.Background(), gwmux, endpoint, dopts)
	if err != nil {
		log.Fatalln("Failed to register gwmux:", err)
	}

	return gwmux
}

func main() {
	log.Printf("grpc or http server is started at: 127.0.0.1:%s \n", grpcAndHttpPort)
	err := RunServer(grpcAndHttpPort)
	if err != nil {
		log.Fatalf("Run Serve err: %v", err)
	}

}

// 测试
// go run main.go
// grpcurl -plaintext -d '{"name":"go"}' localhost:8004 proto.TagService/GetTagList
// curl 'http://127.0.0.1:8004/api/v1/tags?name=Go'
// curl http://127.0.0.1:8004/ping
