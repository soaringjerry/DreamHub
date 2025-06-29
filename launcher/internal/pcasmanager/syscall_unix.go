// +build !windows

package pcasmanager

import (
	"os/exec"
	"syscall"
)

// setProcAttr 设置进程属性（Unix 系统）
func setProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}