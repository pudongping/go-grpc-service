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

type TagServer struct {
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
