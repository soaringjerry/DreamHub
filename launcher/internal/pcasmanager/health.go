package pcasmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/dreamhub-project/dreamhub/launcher/internal/protos"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// HealthStatus 表示健康检查的状态
type HealthStatus string

const (
	HealthStatusUnknown    HealthStatus = "UNKNOWN"
	HealthStatusServing    HealthStatus = "SERVING"
	HealthStatusNotServing HealthStatus = "NOT_SERVING"
)

// HealthCheck 执行健康检查
func (m *Manager) HealthCheck() (HealthStatus, error) {
	// 读取运行时信息获取端口
	info, err := readRuntimeInfo()
	if err != nil {
		return HealthStatusUnknown, fmt.Errorf("无法读取运行时信息: %v", err)
	}

	// 创建 gRPC 连接
	addr := fmt.Sprintf("localhost:%d", info.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return HealthStatusUnknown, fmt.Errorf("无法连接到 PCAS 服务 (%s): %v", addr, err)
	}
	defer conn.Close()

	// 创建健康检查客户端
	client := protos.NewHealthClient(conn)

	// 执行健康检查
	req := &protos.HealthCheckRequest{
		Service: "pcas",
	}
	
	checkCtx, checkCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer checkCancel()
	
	resp, err := client.Check(checkCtx, req)
	if err != nil {
		return HealthStatusUnknown, fmt.Errorf("健康检查失败: %v", err)
	}

	// 转换响应状态
	switch resp.Status {
	case protos.HealthCheckResponse_SERVING:
		return HealthStatusServing, nil
	case protos.HealthCheckResponse_NOT_SERVING:
		return HealthStatusNotServing, nil
	default:
		return HealthStatusUnknown, nil
	}
}

// GetHealthInfo 获取健康状态的详细信息
func (m *Manager) GetHealthInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	// 检查进程状态
	m.mutex.Lock()
	info["process_running"] = m.isRunning
	info["pid"] = m.pid
	m.mutex.Unlock()

	// 尝试读取运行时信息
	if runtimeInfo, err := readRuntimeInfo(); err == nil {
		info["port"] = runtimeInfo.Port
		info["runtime_pid"] = runtimeInfo.PID
	}

	// 执行健康检查
	if status, err := m.HealthCheck(); err == nil {
		info["health_status"] = string(status)
	} else {
		info["health_status"] = "ERROR"
		info["health_error"] = err.Error()
	}

	return info
}