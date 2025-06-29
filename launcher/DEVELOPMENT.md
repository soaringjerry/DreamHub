# DreamHub Launcher 开发环境设置

## 必需的开发工具

### 1. Protocol Buffers 编译器 (protoc)

protoc 是生成 gRPC 代码的核心工具。

**Ubuntu/Debian 安装方式：**
```bash
# 方式一：使用 apt (版本可能较旧)
sudo apt update
sudo apt install -y protobuf-compiler

# 方式二：从 GitHub 下载最新版本
PROTOC_ZIP=protoc-21.12-linux-x86_64.zip
curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v21.12/$PROTOC_ZIP
sudo unzip -o $PROTOC_ZIP -d /usr/local bin/protoc
sudo unzip -o $PROTOC_ZIP -d /usr/local 'include/*'
rm -f $PROTOC_ZIP
```

**macOS 安装方式：**
```bash
brew install protobuf
```

**验证安装：**
```bash
protoc --version
```

### 2. Go 插件

这些插件用于生成 Go 语言的 gRPC 代码。

```bash
# 安装 Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 确保 Go bin 目录在 PATH 中
export PATH="$PATH:$(go env GOPATH)/bin"
```

### 3. 系统依赖（用于系统托盘）

**Ubuntu/Debian：**
```bash
sudo apt-get install gcc libgtk-3-dev libayatana-appindicator3-dev
```

**Fedora：**
```bash
sudo dnf install gcc gtk3-devel libappindicator-gtk3-devel
```

**macOS：**
无需额外依赖

## 生成 gRPC 代码

安装完所有依赖后，运行以下命令生成 gRPC 代码：

```bash
cd launcher/protos
./generate.sh
```

这将生成：
- `health.pb.go` - Protocol Buffers 消息定义
- `health_grpc.pb.go` - gRPC 服务客户端和服务器代码

## 构建项目

### 完整构建（包含系统托盘）
```bash
cd launcher
./build.sh
```

### 仅 CLI 模式（无需系统依赖）
```bash
cd launcher
./build.sh nosystray
```

## 故障排除

1. **protoc: command not found**
   - 确保 protoc 已正确安装
   - 检查 PATH 环境变量

2. **protoc-gen-go: program not found**
   - 运行 `go install` 命令安装插件
   - 确保 `$(go env GOPATH)/bin` 在 PATH 中

3. **系统托盘编译错误**
   - 安装必需的系统库
   - 或使用 `nosystray` 标签构建