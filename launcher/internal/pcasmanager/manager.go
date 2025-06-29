package pcasmanager

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
)

// Manager 负责管理 PCAS 进程的生命周期
type Manager struct {
	cmd            *exec.Cmd
	pid            int
	port           int    // PCAS 使用的端口
	isRunning      bool
	mutex          sync.Mutex
	statusCallback func() // 状态变化回调函数
}

// New 创建一个新的 Manager 实例
func New() *Manager {
	return &Manager{
		isRunning: false,
	}
}

// IsRunning 返回 PCAS 进程是否正在运行
func (m *Manager) IsRunning() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.isRunning
}

// Start 启动 PCAS 进程
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return errors.New("PCAS 进程已经在运行")
	}

	// 获取当前可执行文件的目录
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取可执行文件路径: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	
	// 构造 PCAS 可执行文件路径
	pcasPath := filepath.Join(exeDir, "core", "pcas")
	if runtime.GOOS == "windows" {
		pcasPath += ".exe"
	}

	// 检查文件是否存在
	if _, err := os.Stat(pcasPath); os.IsNotExist(err) {
		return fmt.Errorf("PCAS 可执行文件不存在: %s", pcasPath)
	}

	// 查找可用端口
	port, err := findAvailablePort(50051)
	if err != nil {
		return fmt.Errorf("查找可用端口失败: %v", err)
	}
	m.port = port

	// 创建命令，传递端口参数
	m.cmd = exec.Command(pcasPath, "--grpc-port", fmt.Sprintf("%d", port))
	
	// 设置进程属性
	setProcAttr(m.cmd)

	// 启动进程
	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("启动 PCAS 进程失败: %v", err)
	}

	// 保存进程信息
	m.pid = m.cmd.Process.Pid
	m.isRunning = true

	// 写入运行时信息
	if err := writeRuntimeInfo(m.pid, m.port); err != nil {
		// 写入失败不应该阻止进程启动，但要记录错误
		fmt.Printf("警告: 写入 runtime.json 失败: %v\n", err)
	}

	// 启动 goroutine 监控进程状态
	go m.waitForExit()

	return nil
}

// SetStatusCallback 设置状态变化回调函数
func (m *Manager) SetStatusCallback(callback func()) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.statusCallback = callback
}

// waitForExit 等待进程退出并更新状态
func (m *Manager) waitForExit() {
	// 等待进程退出
	m.cmd.Wait()

	// 更新状态
	m.mutex.Lock()
	m.isRunning = false
	m.cmd = nil
	m.pid = 0
	m.port = 0
	callback := m.statusCallback
	m.mutex.Unlock()

	// 删除运行时信息文件
	if err := removeRuntimeInfo(); err != nil {
		fmt.Printf("警告: 删除 runtime.json 失败: %v\n", err)
	}

	// 触发状态回调
	if callback != nil {
		callback()
	}
}

// Stop 优雅地停止 PCAS 进程
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return errors.New("PCAS 进程未在运行")
	}

	if m.cmd == nil || m.cmd.Process == nil {
		return errors.New("进程引用无效")
	}

	// 发送 SIGTERM 信号以优雅地终止进程
	var err error
	if runtime.GOOS == "windows" {
		// Windows 不支持 SIGTERM，使用 Kill
		// 在实际应用中，可以考虑使用 Windows API 发送 WM_CLOSE 消息
		err = m.cmd.Process.Kill()
	} else {
		// Unix-like 系统使用 SIGTERM
		err = m.cmd.Process.Signal(syscall.SIGTERM)
	}

	if err != nil {
		return fmt.Errorf("停止 PCAS 进程失败: %v", err)
	}

	// 停止成功后也删除运行时信息文件
	if err := removeRuntimeInfo(); err != nil {
		fmt.Printf("警告: 删除 runtime.json 失败: %v\n", err)
	}

	return nil
}

// GetPID 返回当前运行的 PCAS 进程的 PID
func (m *Manager) GetPID() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.pid
}

// GetPort 返回当前 PCAS 使用的端口
func (m *Manager) GetPort() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.port
}