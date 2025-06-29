#!/bin/bash

# 生成 gRPC 代码的脚本
# 需要先安装 protoc: https://grpc.io/docs/protoc-installation/

echo "生成 gRPC 代码..."

# 确保在正确的目录
cd "$(dirname "$0")"

# 生成 Go 代码
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    health.proto

echo "完成!"