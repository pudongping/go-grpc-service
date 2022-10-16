package main

import (
	"context"
	"log"

	pb "github.com/pudongping/go-grpc-service/proto"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	clientConn, _ := GetClientConn(ctx, "localhost:5200", nil)
	defer clientConn.Close()

	// 初始化指定 RPC Proto Service 的客户端实例对象
	tagServiceClient := pb.NewTagServiceClient(clientConn)
	// 发起指定 RPC 方法的调用
	resp, err := tagServiceClient.GetTagList(ctx, &pb.GetTagListRequest{Name: "Go"})

	if err != nil {
		log.Printf("client err ====> %v", err)
	}

	log.Printf("client resp ====> %v", resp)
}

func GetClientConn(ctx context.Context, target string, opts []grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithInsecure())

	// 创建给定目标的客户端连接，另外我们所要请求的服务端是非加密模式的，
	// 因此我们调用了 grpc.WithInsecure 方法禁用了此 ClientConn 的传输安全性验证
	return grpc.DialContext(ctx, target, opts...)
}
