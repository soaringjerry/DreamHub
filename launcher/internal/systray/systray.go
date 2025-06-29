// +build !nosystray

package systray

import (
	"fmt"
	"log"

	"github.com/dreamhub-project/dreamhub/launcher/internal/pcasmanager"
	"github.com/getlantern/systray"
)

var manager *pcasmanager.Manager

func Run() {
	// 创建 PCAS 管理器实例
	manager = pcasmanager.New()
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("DreamHub")
	systray.SetTooltip("DreamHub Launcher")

	statusItem := systray.AddMenuItem("状态: 未启动", "当前 PCAS 核心状态")
	statusItem.Disable()

	startItem := systray.AddMenuItem("启动 PCAS", "启动 PCAS 核心服务")
	stopItem := systray.AddMenuItem("停止 PCAS", "停止 PCAS 核心服务")
	stopItem.Disable() // 初始状态下禁用停止按钮
	
	systray.AddSeparator()
	
	quitItem := systray.AddMenuItem("退出启动器", "退出 DreamHub Launcher")

	// 更新状态的辅助函数
	updateStatus := func() {
		if manager.IsRunning() {
			port := manager.GetPort()
			if port > 0 {
				statusItem.SetTitle(fmt.Sprintf("状态: 运行中 (PID: %d, 端口: %d)", manager.GetPID(), port))
			} else {
				statusItem.SetTitle(fmt.Sprintf("状态: 运行中 (PID: %d)", manager.GetPID()))
			}
			startItem.Disable()
			stopItem.Enable()
		} else {
			statusItem.SetTitle("状态: 未启动")
			startItem.Enable()
			stopItem.Disable()
		}
	}

	// 设置状态回调，以便在进程意外退出时更新 UI
	manager.SetStatusCallback(func() {
		log.Println("PCAS 进程已退出")
		updateStatus()
	})

	go func() {
		for {
			select {
			case <-startItem.ClickedCh:
				log.Println("启动按钮被点击")
				if err := manager.Start(); err != nil {
					log.Printf("启动 PCAS 失败: %v", err)
				} else {
					log.Printf("PCAS 已启动，PID: %d", manager.GetPID())
					updateStatus()
				}
			case <-stopItem.ClickedCh:
				log.Println("停止按钮被点击")
				if err := manager.Stop(); err != nil {
					log.Printf("停止 PCAS 失败: %v", err)
				} else {
					log.Println("PCAS 已停止")
					updateStatus()
				}
			case <-quitItem.ClickedCh:
				// 退出前确保停止 PCAS
				if manager.IsRunning() {
					log.Println("退出前停止 PCAS...")
					manager.Stop()
				}
				systray.Quit()
			}
		}
	}()
}

func onExit() {
	// 清理工作：确保 PCAS 进程被停止
	if manager != nil && manager.IsRunning() {
		log.Println("清理：停止 PCAS 进程...")
		manager.Stop()
	}
}