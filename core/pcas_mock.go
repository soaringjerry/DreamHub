package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Println("PCAS Mock 进程已启动")
	fmt.Printf("PID: %d\n", os.Getpid())

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// 模拟持续运行的服务
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("PCAS Mock 正在运行...")
		case sig := <-sigChan:
			log.Printf("收到信号: %v，准备优雅退出", sig)
			// 模拟清理工作
			time.Sleep(1 * time.Second)
			log.Println("PCAS Mock 已优雅退出")
			os.Exit(0)
		}
	}
}