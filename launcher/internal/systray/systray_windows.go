// +build windows,!nosystray

package systray

import (
	"syscall"
	"unsafe"
	
	"github.com/dreamhub-project/dreamhub/launcher/internal/pcasmanager"
	"github.com/getlantern/systray"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	messageBox = user32.NewProc("MessageBoxW")
)

var manager *pcasmanager.Manager

func Run() {
	// 创建 PCAS 管理器实例
	manager = pcasmanager.New()
	
	// 尝试运行系统托盘
	defer func() {
		if r := recover(); r != nil {
			// 如果系统托盘初始化失败，显示错误消息
			showErrorMessage("DreamHub Launcher", 
				"无法初始化系统托盘。\n\n" +
				"请确保：\n" +
				"1. Windows 资源管理器正在运行\n" +
				"2. 系统托盘没有被禁用\n\n" +
				"您可以使用命令行模式：dreamhub.exe status")
		}
	}()
	
	systray.Run(onReady, onExit)
}

func showErrorMessage(title, message string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)
	messageBox.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		0x10, // MB_ICONERROR
	)
}