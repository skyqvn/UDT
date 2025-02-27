package main

import (
	"os"
	"syscall"
	"unsafe"
)

// 调用 Windows API 以管理员权限启动程序
func runAsAdmin(exePath, args string) error {
	// 加载 kernel32.dll 库
	kernel32 := syscall.NewLazyDLL("shell32.dll")
	// 获取 ShellExecuteW 函数地址
	shellExecute := kernel32.NewProc("ShellExecuteW")
	
	// 将字符串转换为 UTF-16 指针
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exePath)
	params, _ := syscall.UTF16PtrFromString(args)
	directory, _ := syscall.UTF16PtrFromString("")
	
	// 调用 ShellExecuteW 函数
	ret, _, err := shellExecute.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(params)),
		uintptr(unsafe.Pointer(directory)),
		1, // SW_SHOWNORMAL
	)
	
	// 检查返回值
	if ret <= 32 {
		return err
	}
	return nil
}

func isAdmin() bool {
	_, err := os.Open(`\\.\PHYSICALDRIVE0`)
	return err == nil
}
