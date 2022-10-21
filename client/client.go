package main

import (
	"context"
	"log"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pudongping/go-grpc-service/global"
	"github.com/pudongping/go-grpc-service/internal/middleware"
	"github.com/pudongping/go-grpc-service/pkg/tracer"
	pb "github.com/pudongping/go-grpc-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func init() {
	err := setupTracer()
	if err != nil {
		log.Fatalf("init.setupTracer err: %v", err)
	}
}

type Auth struct {
	AppKey    string
	AppSecret string
}

// 类似于设置 header 信息
func (a *Auth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"app_key":    a.AppKey,
		"app_secret": a.AppSecret,
	}, nil
}

// 是否开启 tls
func (a *Auth) RequireTransportSecurity() bool {
	return false
}

func main() {
	ctx := context.Background()

	// 添加授权认证信息
	auth := Auth{
		AppKey:    "alex",
		AppSecret: "never_give_up",
	}

	// 客户端添加 metadata 数据
	// 在新增 metadata 信息时，务必使用 Append 类别的方法，
	// 否则如果直接 New 一个全新的 md，将会导致原有的 metadata 信息丢失（除非你确定你希望得到这样的结果）
	newCtx := metadata.AppendToOutgoingContext(ctx, "name", "alex")
	// NewIncomingContext：创建一个附加了所传入的 md 新上下文，仅供自身的 gRPC 服务端内部使用。
	// NewOutgoingContext：创建一个附加了传出 md 的新上下文，可供外部的 gRPC 客户端、服务端使用
	// newCtx := metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"name": "alex"}))

	clientConn, err := GetClientConn(newCtx, "localhost:8004", []grpc.DialOption{
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(
				middleware.UnaryContextTimeout(),
				middleware.ClientTracing(), // 链路追踪
			),
		),
		grpc.WithPerRPCCredentials(&auth), // 做自定义认证
	})
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	defer clientConn.Close()

	// 初始化指定 RPC Proto Service 的客户端实例对象
	tagServiceClient := pb.NewTagServiceClient(clientConn)
	// 发起指定 RPC 方法的调用
	resp, err := tagServiceClient.GetTagList(newCtx, &pb.GetTagListRequest{Name: "Go"})

	if err != nil {
		log.Printf("client err ====> %v", err)
	}

	log.Printf("client resp ====> %v", resp)
}

func GetClientConn(ctx context.Context, target string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// 创建给定目标的客户端连接，另外我们所要请求的服务端是非加密模式的，
	// 因此我们调用了 grpc.WithInsecure 方法禁用了此 ClientConn 的传输安全性验证
	return grpc.DialContext(ctx, target, opts...)
}

func setupTracer() error {
	var err error
	jaegerTracer, _, err := tracer.NewJaegerTracer("article-service", "127.0.0.1:6831")
	if err != nil {
		return err
	}
	global.Tracer = jaegerTracer
	return nil
}
