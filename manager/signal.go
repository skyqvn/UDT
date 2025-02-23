package main

import (
	"os"
	"syscall"
)

// 加载 kernel32.dll 动态链接库
var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
)

func sendExitSignal(pid int) error {
	// 根据PID获取进程对象
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	
	// 尝试终止进程
	err = process.Kill()
	if err != nil {
		return err
	}
	
	// 等待进程退出
	_, err = process.Wait()
	if err != nil {
		return err
	}
	
	return nil
}
