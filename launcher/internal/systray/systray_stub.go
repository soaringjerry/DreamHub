// +build nosystray

package systray

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"github.com/dreamhub-project/dreamhub/launcher/internal/pcasmanager"
)

func Run() {
	fmt.Println("===========================================")
	fmt.Println("DreamHub Launcher v0.3.0")
	fmt.Println("===========================================")
	fmt.Println()
	
	manager := pcasmanager.New()
	
	// 自动启动 PCAS
	fmt.Println("正在启动 PCAS 服务...")
	if err := manager.Start(); err != nil {
		fmt.Printf("启动失败: %v\n", err)
		fmt.Println("\n按任意键退出...")
		fmt.Scanln()
		os.Exit(1)
	}
	
	fmt.Println("PCAS 服务已成功启动！")
	
	// 等待一下让服务完全启动
	time.Sleep(2 * time.Second)
	
	// 显示状态
	info := manager.GetHealthInfo()
	fmt.Println("\n===== 服务状态 =====")
	if running, ok := info["process_running"].(bool); ok {
		fmt.Printf("进程运行: %v\n", running)
	}
	if pid, ok := info["pid"].(int); ok && pid > 0 {
		fmt.Printf("进程 PID: %d\n", pid)
	}
	if port, ok := info["port"].(int); ok && port > 0 {
		fmt.Printf("服务端口: %d\n", port)
		fmt.Printf("\n您可以使用以下命令测试服务:\n")
		fmt.Printf("pcasctl.exe --port %d search \"test\"\n", port)
	} else {
		fmt.Println("警告: 无法获取服务端口信息")
		fmt.Println("PCAS 进程可能已经退出或未正确启动")
	}
	
	fmt.Println("\nDreamHub 正在后台运行。")
	fmt.Println("按 Ctrl+C 停止服务并退出。")
	
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// 等待退出信号
	<-sigChan
	
	fmt.Println("\n正在停止服务...")
	manager.Stop()
	fmt.Println("服务已停止。再见！")
}