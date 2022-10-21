package server

import (
	"context"
	"encoding/json"
	"log"

	"github.com/pudongping/go-grpc-service/pkg/bapi"
	"github.com/pudongping/go-grpc-service/pkg/errcode"
	pb "github.com/pudongping/go-grpc-service/proto"
	"google.golang.org/grpc/metadata"
)

type Auth struct {
}

func (a *Auth) GetAppKey() string {
	return "alex"
}

func (a *Auth) GetAppSecret() string {
	return "never_give_up"
}

func (a *Auth) Check(ctx context.Context) error {
	// FromIncomingContext：读取 metadata，仅供自身的 gRPC 服务端内部使用。
	md, _ := metadata.FromIncomingContext(ctx)

	var appKey, appSecret string
	if value, ok := md["app_key"]; ok {
		appKey = value[0]
	}
	if value, ok := md["app_secret"]; ok {
		appSecret = value[0]
	}

	if appKey != a.GetAppKey() || appSecret != a.GetAppSecret() {
		return errcode.TogRPCError(errcode.Unauthorized)
	}

	return nil
}

type TagServer struct {
	auth *Auth
}

func NewTagServer() *TagServer {
	return &TagServer{}
}

func (t *TagServer) GetTagList(ctx context.Context, r *pb.GetTagListRequest) (*pb.GetTagListResponse, error) {
	// 读取客户端传递过来的 metadata 数据
	// FromIncomingContext：读取 metadata，仅供自身的 gRPC 服务端内部使用。
	// FromOutgoingContext：读取 metadata，可供外部的 gRPC 客户端、服务端使用。
	md, _ := metadata.FromIncomingContext(ctx)
	log.Printf("获取类似于 header 头的信息为 ====> %+v \n", md)

	// 授权验证
	if err := t.auth.Check(ctx); err != nil {
		return nil, err
	}

	api := bapi.NewAPI("http://127.0.0.1:8000")
	body, err := api.GetTagList(ctx, r.GetName())
	if err != nil {
		return nil, errcode.TogRPCError(errcode.ErrorGetTagListFail)
	}

	tagList := pb.GetTagListResponse{}
	err = json.Unmarshal(body, &tagList)
	if err != nil {
		return nil, errcode.TogRPCError(errcode.Fail)
	}

	return &tagList, nil
}
