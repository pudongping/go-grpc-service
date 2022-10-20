#!/usr/bin/env bash
# generate proto pb files

# 没有使用 grpc-gateway 插件前
#protoc --go_out=plugins=grpc:. ./*.proto

# 需要使用 grpc-gateway 插件和 swagger 插件时
protoc -I. \
--go_out=plugins=grpc:. \
--grpc-gateway_out=logtostderr=true:. \
--swagger_out=logtostderr=true:. \
./*.proto
