package pcasmanager

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

// RuntimeInfo 存储运行时信息
type RuntimeInfo struct {
	PID  int `json:"pid"`
	Port int `json:"port"`
}

// findAvailablePort 从指定端口开始查找可用的端口
func findAvailablePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			// 端口被占用，继续尝试下一个
			continue
		}
		// 端口可用，立即关闭监听器
		listener.Close()
		return port, nil
	}
	return 0, fmt.Errorf("无法在 %d-%d 范围内找到可用端口", startPort, startPort+99)
}

// writeRuntimeInfo 将运行时信息写入 runtime.json
func writeRuntimeInfo(pid, port int) error {
	// 确保 data 目录存在
	dataDir := filepath.Join("..", "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("创建 data 目录失败: %v", err)
	}

	// 准备运行时信息
	info := RuntimeInfo{
		PID:  pid,
		Port: port,
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化运行时信息失败: %v", err)
	}

	// 写入文件
	runtimePath := filepath.Join(dataDir, "runtime.json")
	if err := os.WriteFile(runtimePath, data, 0644); err != nil {
		return fmt.Errorf("写入 runtime.json 失败: %v", err)
	}

	return nil
}

// readRuntimeInfo 从 runtime.json 读取运行时信息
func readRuntimeInfo() (*RuntimeInfo, error) {
	runtimePath := filepath.Join("..", "data", "runtime.json")
	
	// 读取文件
	data, err := os.ReadFile(runtimePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("runtime.json 不存在")
		}
		return nil, fmt.Errorf("读取 runtime.json 失败: %v", err)
	}

	// 解析 JSON
	var info RuntimeInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("解析 runtime.json 失败: %v", err)
	}

	return &info, nil
}

// removeRuntimeInfo 删除 runtime.json 文件
func removeRuntimeInfo() error {
	runtimePath := filepath.Join("..", "data", "runtime.json")
	
	// 如果文件不存在，不算错误
	if err := os.Remove(runtimePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除 runtime.json 失败: %v", err)
	}
	
	return nil
}