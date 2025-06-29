// +build windows

package pcasmanager

import (
	"os/exec"
	"syscall"
)

// setProcAttr 设置进程属性（Windows 系统）
func setProcAttr(cmd *exec.Cmd) {
	// Windows 不需要设置进程组
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true, // 隐藏控制台窗口
	}
}