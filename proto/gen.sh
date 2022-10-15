#!/usr/bin/env bash
# generate proto pb files
protoc --go_out=plugins=grpc:. ./*.proto