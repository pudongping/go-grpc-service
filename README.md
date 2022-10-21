# go-grpc-service
学习 grpc 写的一部分服务代码

## swagger 文档

访问 http://127.0.0.1:8004/swagger-ui/ 即可访问 swagger 面板  
访问 http://127.0.0.1:8004/swagger/tag.swagger.json 即可访问到自己的接口文档

## 链路追踪

### docker 安装 jaeger

```shell
docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.28
```

安装完毕后，启动服务，随便调用一个接口，然后浏览器访问 http://127.0.0.1:16686/search 即可查看

## 特殊文件说明

文件 | 说明
--- | ---
根目录下 main1.go | 监听两个不同的端口，一个用作 http 服务，一个用作 grpc 服务
根目录下 main2.go | 一个端口上兼容多种协议，一个连接可以是 gRPC 或 HTTP 但不能同时是两者
根目录下 main3.go | 使用 grpc-gateway 提供同端口同方法提供双流量支持

## 项目启动

### 启动服务

```shell
# 默认启动了链路追踪，因此需要使用 docker 开启 jaeger
go run main.go
```

### 测试

```shell
# 查看所有的标签
curl 'http://127.0.0.1:8004/api/v1/tags'  
# {"list":[{"id":"1","name":"Go","state":1},{"id":"2","name":"PHP","state":1},{"id":"3","name":"Rust","state":1}],"pager":{"page":"1","pageSize":"10","totalRows":"3"}}

# 查看指定标签
curl 'http://127.0.0.1:8004/api/v1/tags?name=Go'
{"list":[{"id":"1","name":"Go","state":1}],"pager":{"page":"1","pageSize":"10","totalRows":"1"}}

# 心跳检测
curl 'http://127.0.0.1:8004/ping'
# pong

# 测试访问 grpc
grpcurl -plaintext -d '{"name":"go"}' localhost:8004 proto.TagService/GetTagList
#{
#  "list": [
#    {
#      "id": "1",
#      "name": "go",
#      "state": 1
#    }
#  ],
#  "pager": {
#    "page": "1",
#    "pageSize": "10",
#    "totalRows": "1"
#  }
#}

# 不带任何参数访问 grpc
grpcurl -plaintext localhost:8004 proto.TagService/GetTagList
#{
#  "list": [
#    {
#      "id": "1",
#      "name": "Go",
#      "state": 1
#    },
#    {
#      "id": "2",
#      "name": "PHP",
#      "state": 1
#    },
#    {
#      "id": "3",
#      "name": "Rust",
#      "state": 1
#    }
#  ],
#  "pager": {
#    "page": "1",
#    "pageSize": "10",
#    "totalRows": "3"
#  }
#}

```

可能会输出以下类似内容

```shell
2022/10/22 00:12:54 grpc or http server is started at: 127.0.0.1:8004 
2022/10/22 00:13:01 access request log: method: /proto.TagService/GetTagList, begin_time: 1666368781, request: 
2022/10/22 00:13:01 获取类似于 header 头的信息为 ====> map[:authority:[0.0.0.0:8004] content-type:[application/grpc] grpcgateway-accept:[*/*] grpcgateway-user-agent:[curl/7.64.1] user-agent:[grpc-go/1.50.1] x-forwarded-for:[127.0.0.1] x-forwarded-host:[127.0.0.1:8004]] 
2022/10/22 00:13:01 access response log: method: /proto.TagService/GetTagList, begin_time: 1666368781, end_time: 1666368781, response: list:{id:1 name:"Go" state:1} list:{id:2 name:"PHP" state:1} list:{id:3 name:"Rust" state:1} pager:{page:1 page_size:10 total_rows:3}
2022/10/22 00:13:07 access request log: method: /proto.TagService/GetTagList, begin_time: 1666368787, request: name:"Go"
2022/10/22 00:13:07 获取类似于 header 头的信息为 ====> map[:authority:[0.0.0.0:8004] content-type:[application/grpc] grpcgateway-accept:[*/*] grpcgateway-user-agent:[curl/7.64.1] user-agent:[grpc-go/1.50.1] x-forwarded-for:[127.0.0.1] x-forwarded-host:[127.0.0.1:8004]] 
2022/10/22 00:13:07 access response log: method: /proto.TagService/GetTagList, begin_time: 1666368787, end_time: 1666368787, response: list:{id:1 name:"Go" state:1} pager:{page:1 page_size:10 total_rows:1}
2022/10/22 00:14:04 access request log: method: /proto.TagService/GetTagList, begin_time: 1666368844, request: name:"go"
2022/10/22 00:14:04 获取类似于 header 头的信息为 ====> map[:authority:[localhost:8004] content-type:[application/grpc] user-agent:[grpcurl/dev-build (no version set) grpc-go/1.48.0]] 
2022/10/22 00:14:04 access response log: method: /proto.TagService/GetTagList, begin_time: 1666368844, end_time: 1666368844, response: list:{id:1 name:"go" state:1} pager:{page:1 page_size:10 total_rows:1}
2022/10/22 00:14:32 access request log: method: /proto.TagService/GetTagList, begin_time: 1666368872, request: 
2022/10/22 00:14:32 获取类似于 header 头的信息为 ====> map[:authority:[localhost:8004] content-type:[application/grpc] user-agent:[grpcurl/dev-build (no version set) grpc-go/1.48.0]] 
2022/10/22 00:14:32 access response log: method: /proto.TagService/GetTagList, begin_time: 1666368872, end_time: 1666368872, response: list:{id:1 name:"Go" state:1} list:{id:2 name:"PHP" state:1} list:{id:3 name:"Rust" state:1} pager:{page:1 page_size:10 total_rows:3}
```

### grpc 客户端代码访问

```shell
go run client/client.go
# 2022/10/22 00:20:31 client resp ====> list:{id:1 name:"Go" state:1} pager:{page:1 page_size:10 total_rows:1}
```

可能会输出以下类似内容

```shell
2022/10/22 00:20:31 access request log: method: /proto.TagService/GetTagList, begin_time: 1666369231, request: name:"Go"
2022/10/22 00:20:31 获取类似于 header 头的信息为 ====> map[:authority:[localhost:8004] app_key:[alex] app_secret:[never_give_up] content-type:[application/grpc] name:[alex] uber-trace-id:[198e5a4758afdcac:198e5a4758afdcac:0000000000000000:1] user-agent:[grpc-go/1.50.1]] 
2022/10/22 00:20:31 access response log: method: /proto.TagService/GetTagList, begin_time: 1666369231, end_time: 1666369231, response: list:{id:1 name:"Go" state:1} pager:{page:1 page_size:10 total_rows:1}
```