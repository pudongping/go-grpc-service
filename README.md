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