package main

import (
	"fmt"
	"golang.org/x/sys/windows"
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
		case "install":
			printLogo()
			setAutoStart(true)
			fmt.Println("安装成功！\r")
			fmt.Println("按Enter键结束安装程序\r")
			fmt.Scanln()
		case "uninstall":
			printLogo()
			setAutoStart(false)
			fmt.Println("卸载成功！")
			fmt.Println("按Enter键结束安装程序")
			fmt.Scanln()
		}
	} else {
		fmt.Errorf("无管理员权限")
		os.Exit(1)
	}
}

// 获取 Windows 临时目录
func getTempDir() string {
	// 全局单例运行
	return `C:\Windows\Temp`
	// 用户单例运行
	// s, ok := os.LookupEnv("TEMP")
	// if ok {
	// 	return s
	// }
	// return os.TempDir()
}

func stop() {
	pid := getPid()
	if pid != -1 {
		err := sendExitSignal(pid)
		if err != nil {
			fmt.Printf("关闭进程失败: %v\n", err)
			return
		}
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
	fmt.Println("已成功启动UDT")
}

func reboot() {
	stop()
	start()
}

func edit() {
	cmd := exec.Command("vim", "./config.yaml")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("\nVim编辑失败: %v\n", err)
	}
	fmt.Println("编辑成功")
	reboot()
}
