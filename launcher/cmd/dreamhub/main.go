package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/dreamhub-project/dreamhub/launcher/internal/pcasmanager"
	"github.com/dreamhub-project/dreamhub/launcher/internal/systray"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dreamhub",
	Short: "DreamHub Launcher - 管理 PCAS 核心的桌面启动器",
	Long:  `DreamHub Launcher 是一个用于管理 PCAS (Personal Cloud AI System) 核心的桌面启动器。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("DreamHub Launcher 正在启动...")
		fmt.Println("程序将在系统托盘运行，请查看屏幕右下角的托盘图标。")
		systray.Run()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "检查 PCAS 核心状态",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("正在检查 PCAS 服务状态...")
		fmt.Println("=====================================")
		
		// 创建管理器实例
		manager := pcasmanager.New()
		
		// 获取详细健康信息
		info := manager.GetHealthInfo()
		
		// 显示进程状态
		if running, ok := info["process_running"].(bool); ok && running {
			fmt.Printf("进程状态: 运行中\n")
			if pid, ok := info["pid"].(int); ok && pid > 0 {
				fmt.Printf("进程 PID: %d\n", pid)
			}
		} else {
			fmt.Printf("进程状态: 未运行\n")
		}
		
		// 显示端口信息
		if port, ok := info["port"].(int); ok && port > 0 {
			fmt.Printf("服务端口: %d\n", port)
		}
		
		// 显示健康状态
		if status, ok := info["health_status"].(string); ok {
			fmt.Printf("健康状态: %s\n", status)
			
			// 如果有错误，显示错误信息
			if status == "ERROR" {
				if errMsg, ok := info["health_error"].(string); ok {
					fmt.Printf("错误信息: %s\n", errMsg)
				}
			}
		}
		
		fmt.Println("=====================================")
		
		// 根据状态给出总结
		if status, ok := info["health_status"].(string); ok && status == "SERVING" {
			fmt.Println("✅ PCAS 服务正常运行")
		} else if running, ok := info["process_running"].(bool); ok && running {
			fmt.Println("⚠️  PCAS 进程已启动，但服务尚未就绪")
		} else {
			fmt.Println("❌ PCAS 服务未运行")
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func main() {
	// Windows 下的特殊处理
	if len(os.Args) == 1 && runtime.GOOS == "windows" {
		// 如果是 Windows 且没有参数，添加一个小延迟确保能看到输出
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("程序发生错误: %v\n", r)
				fmt.Println("按任意键退出...")
				fmt.Scanln()
			}
		}()
	}
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		if runtime.GOOS == "windows" && len(os.Args) == 1 {
			fmt.Println("按任意键退出...")
			fmt.Scanln()
		}
		os.Exit(1)
	}
}