package main

import (
	"fmt"
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)

// 调用 Windows API 以管理员权限启动程序
func runAsAdmin(exePath, args string) error {
	// 将字符串转换为 UTF-16 指针
	verb, _ := syscall.UTF16PtrFromString("runas")
	file, _ := syscall.UTF16PtrFromString(exePath)
	params, _ := syscall.UTF16PtrFromString(args)
	directory, _ := syscall.UTF16PtrFromString("")

	// 调用 ShellExecuteW 函数
	ret, _, err := ShellExecuteW.Call(
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
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid)
	if err != nil {
		fmt.Printf("Failed to allocate and initialize SID: %v\n", err)
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.GetCurrentProcessToken()
	member, err := token.IsMember(sid)
	return member && err == nil
}
