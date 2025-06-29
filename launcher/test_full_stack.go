// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/dreamhub-project/dreamhub/launcher/pcasmanager"
)

func main() {
	fmt.Println("=== DreamHub 全栈测试 ===")
	fmt.Println()

	// 创建管理器
	manager := pcasmanager.New()
	
	// 设置状态回调
	manager.SetStatusCallback(func() {
		fmt.Println("[回调] PCAS 进程已退出")
	})

	// 1. 测试初始状态
	fmt.Println("1. 检查初始状态...")
	runStatusCommand()
	fmt.Println()

	// 2. 启动 PCAS
	fmt.Println("2. 启动 PCAS...")
	if err := manager.Start(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
	fmt.Printf("启动成功! PID: %d, 端口: %d\n", manager.GetPID(), manager.GetPort())
	fmt.Println()

	// 等待服务就绪
	fmt.Println("等待服务就绪...")
	time.Sleep(2 * time.Second)

	// 3. 检查运行状态
	fmt.Println("3. 检查运行状态...")
	runStatusCommand()
	fmt.Println()

	// 4. 执行健康检查
	fmt.Println("4. 执行健康检查...")
	status, err := manager.HealthCheck()
	if err != nil {
		fmt.Printf("健康检查失败: %v\n", err)
	} else {
		fmt.Printf("健康状态: %s\n", status)
	}
	fmt.Println()

	// 5. 检查 runtime.json
	fmt.Println("5. 检查 runtime.json 文件...")
	if _, err := os.Stat("../data/runtime.json"); err == nil {
		data, _ := os.ReadFile("../data/runtime.json")
		fmt.Printf("runtime.json 内容:\n%s\n", string(data))
	} else {
		fmt.Println("runtime.json 不存在")
	}
	fmt.Println()

	// 6. 停止 PCAS
	fmt.Println("6. 停止 PCAS...")
	if err := manager.Stop(); err != nil {
		log.Fatalf("停止失败: %v", err)
	}
	fmt.Println("停止成功!")
	
	// 等待清理完成
	time.Sleep(2 * time.Second)

	// 7. 验证清理
	fmt.Println("7. 验证清理...")
	if _, err := os.Stat("../data/runtime.json"); os.IsNotExist(err) {
		fmt.Println("✅ runtime.json 已被正确删除")
	} else {
		fmt.Println("❌ runtime.json 仍然存在")
	}
	
	// 8. 最终状态检查
	fmt.Println()
	fmt.Println("8. 最终状态检查...")
	runStatusCommand()

	fmt.Println()
	fmt.Println("=== 测试完成 ===")
}

func runStatusCommand() {
	cmd := exec.Command("./dreamhub", "status")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}