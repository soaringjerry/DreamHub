package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dreamhub-project/dreamhub/core/protos"
	"google.golang.org/grpc"
)

// healthServer 实现健康检查服务
type healthServer struct {
	protos.UnimplementedHealthServer
}

// Check 实现健康检查方法
func (s *healthServer) Check(ctx context.Context, req *protos.HealthCheckRequest) (*protos.HealthCheckResponse, error) {
	log.Printf("收到健康检查请求: service=%s", req.Service)
	return &protos.HealthCheckResponse{
		Status: protos.HealthCheckResponse_SERVING,
	}, nil
}

func main() {
	// 解析命令行参数
	var grpcPort int
	flag.IntVar(&grpcPort, "grpc-port", 50051, "gRPC 服务端口")
	flag.Parse()

	log.Printf("PCAS Mock (带 gRPC) 进程已启动")
	fmt.Printf("PID: %d\n", os.Getpid())
	fmt.Printf("gRPC 端口: %d\n", grpcPort)

	// 创建 gRPC 服务器
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		log.Fatalf("无法监听端口 %d: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	protos.RegisterHealthServer(grpcServer, &healthServer{})

	// 启动 gRPC 服务
	go func() {
		log.Printf("gRPC 服务正在监听端口 %d", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC 服务失败: %v", err)
		}
	}()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// 模拟持续运行的服务
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("PCAS Mock 正在运行...")
		case sig := <-sigChan:
			log.Printf("收到信号: %v，准备优雅退出", sig)
			// 停止 gRPC 服务器
			grpcServer.GracefulStop()
			// 模拟清理工作
			time.Sleep(1 * time.Second)
			log.Println("PCAS Mock 已优雅退出")
			os.Exit(0)
		}
	}
}