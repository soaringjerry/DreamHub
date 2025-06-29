// +build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dreamhub-project/dreamhub/launcher/pcasmanager"
)

func main() {
	fmt.Println("开始测试 PCAS Manager...")

	// 创建管理器
	manager := pcasmanager.New()
	
	// 设置状态回调
	manager.SetStatusCallback(func() {
		fmt.Println("[回调] PCAS 进程已退出")
	})

	// 测试启动
	fmt.Println("\n1. 测试启动 PCAS...")
	if err := manager.Start(); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
	fmt.Printf("启动成功! PID: %d\n", manager.GetPID())
	fmt.Printf("运行状态: %v\n", manager.IsRunning())

	// 等待几秒
	fmt.Println("\n等待 3 秒...")
	time.Sleep(3 * time.Second)

	// 测试重复启动
	fmt.Println("\n2. 测试重复启动（应该失败）...")
	if err := manager.Start(); err != nil {
		fmt.Printf("预期的错误: %v\n", err)
	}

	// 测试停止
	fmt.Println("\n3. 测试停止 PCAS...")
	if err := manager.Stop(); err != nil {
		log.Fatalf("停止失败: %v", err)
	}
	fmt.Println("停止成功!")
	
	// 等待进程完全退出
	time.Sleep(2 * time.Second)
	fmt.Printf("运行状态: %v\n", manager.IsRunning())

	// 测试重新启动
	fmt.Println("\n4. 测试重新启动...")
	if err := manager.Start(); err != nil {
		log.Fatalf("重新启动失败: %v", err)
	}
	fmt.Printf("重新启动成功! PID: %d\n", manager.GetPID())

	// 最终停止
	fmt.Println("\n5. 最终停止...")
	if err := manager.Stop(); err != nil {
		log.Fatalf("最终停止失败: %v", err)
	}

	fmt.Println("\n测试完成!")
}