package main

import (
	"atomicgo.dev/cursor"
	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func getPid() int {
	// 构建锁文件的完整路径
	lockFile := filepath.Join(getTempDir(), "UDT", "app.lock")
	os.Remove(lockFile)
	content, err := os.ReadFile(lockFile)
	if err != nil {
		return -1
	}
	pidStr := strings.TrimSpace(string(content))
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return -1
	}
	return pid
}

func execute(command string) {
	if isAdmin() {
		switch command {
		case InstallCurrentUser:
			fmt.Print("\x1b[2J\x1b[H")
			printLogo()
			setAutoStart(true, registry.CURRENT_USER)
			fmt.Println("安装成功！\r")
			pause()
		case InstallAllUsers:
			fmt.Print("\x1b[2J\x1b[H")
			printLogo()
			setAutoStart(true, registry.LOCAL_MACHINE)
			fmt.Println("安装成功！\r")
			pause()
		case Uninstall:
			fmt.Print("\x1b[2J\x1b[H")
			printLogo()
			setAutoStart(false, registry.CURRENT_USER)
			setAutoStart(false, registry.LOCAL_MACHINE)
			fmt.Println("卸载成功！\r")
			pause()
		}
	} else {
		fmt.Errorf("无管理员权限")
		pause()
		os.Exit(1)
	}
}

// 获取 Windows 临时目录
func getTempDir() string {
	return os.TempDir()
}

func stop() {
	pid := getPid()
	if pid != -1 {
		err := sendExitSignal(pid)
		if err != nil {
			fmt.Printf("关闭进程失败: %v\r\n", err)
			return
		}
		lockFile := filepath.Join(getTempDir(), "UDT", "app.lock")
		os.Remove(lockFile)
		fmt.Println("已停止运行UDT\r")
		return
	}
	fmt.Println("UDT未运行\r")
}

func start() {
	// cmd := exec.Command("cmd", "/C", "start", "/B", "./udt.exe")
	cmd := exec.Command("./udt.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP,
	}
	cmd.Start()
	fmt.Println("已成功启动UDT\r")
}

func reboot() {
	stop()
	start()
}

func edit() {
	cmd := exec.Command("notepad", "./config.yaml")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Vim编辑失败: %v\r\n", err)
	}
	fmt.Println("编辑成功\r")
	reboot()
}

func pause() {
	cursor.Show()
	fmt.Print("按任意键继续……")
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		return true, nil
	})
}
