package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"path"
	"strings"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pudongping/go-grpc-service/internal/middleware"
	"github.com/pudongping/go-grpc-service/pkg/swagger"
	pb "github.com/pudongping/go-grpc-service/proto"
	"github.com/pudongping/go-grpc-service/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var grpcAndHttpPort string

func init() {
	flag.StringVar(&grpcAndHttpPort, "grpc_and_http_port", "8004", "同端口下同 rpc 方法提供 gRPC 和 HTTP 双流量访问支持端口")
	flag.Parse()
}

type httpError struct {
	Code    int32  `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func grpcGatewayError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	httpError := httpError{
		Code:    int32(s.Code()),
		Message: s.Message(),
	}
	details := s.Details()
	for _, detail := range details {
		if v, ok := detail.(*pb.Error); ok {
			httpError.Code = v.Code
			httpError.Message = v.Message
		}
	}

	resp, _ := json.Marshal(httpError)
	w.Header().Set("Content-Type", marshaler.ContentType("application/json"))
	w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
	_, _ = w.Write(resp)
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
	// lis, err := net.Listen("tcp", ":"+port)
	// if err != nil {
	// 	log.Fatalln("Failed to listen:", err)
	// }
	// s := &http.Server{
	// 	Addr:    ":" + port,
	// 	Handler: grpcHandlerFunc(grpcS, httpMux), // 请求的统一入口
	// }
	// return s.Serve(lis)
	// ABC====>>>

	// 和以上 ABC 区域代码一致
	return http.ListenAndServe(":"+port, grpcHandlerFunc(grpcS, httpMux))
}

func runHttpServer() *http.ServeMux {
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`pong`))
	})

	prefix := "/swagger-ui/"
	fileServer := http.FileServer(&assetfs.AssetFS{
		Asset:    swagger.Asset,
		AssetDir: swagger.AssetDir,
		Prefix:   "third_party/swagger-ui",
	})

	httpMux.Handle(prefix, http.StripPrefix(prefix, fileServer))
	httpMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "swagger.json") {
			http.NotFound(w, r)
			return
		}

		p := strings.TrimPrefix(r.URL.Path, "/swagger/")
		p = path.Join("proto", p)

		http.ServeFile(w, r, p)
	})

	return httpMux
}

func runGrpcServer() *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			middleware.AccessLog, // 访问日志
		)),
	}

	s := grpc.NewServer(opts...)
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
	// 实际上在调用这类 RegisterXXXXHandlerFromEndpoint 注册方法时，主要是进行 gRPC 连接的创建和管控，
	// 它在内部就已经调用了 grpc.Dial 对 gRPC Server 进行拨号连接，并保持住了一个 Conn 便于后续的 HTTP/1/1 调用转发。
	// 另外在关闭连接的处理上，处理的也比较的稳健，统一都是放到 defer 中进行关闭，又或者根据 context 的上下文来控制连接的关闭时间
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
