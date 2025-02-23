package main

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"path/filepath"
)

// 新增注册表操作函数
func setAutoStart(enabled bool) error {
	// 获取可执行文件路径
	exePath, err := filepath.Abs("./udt.exe")
	if err != nil {
		return fmt.Errorf("获取UDT.exe路径失败: %v", err)
	}
	
	// 打开注册表项
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.ALL_ACCESS,
	)
	if err != nil {
		return fmt.Errorf("打开注册表项失败: %v", err)
	}
	defer key.Close()
	
	if enabled {
		// 设置字符串值
		err = key.SetStringValue("UDT", exePath)
	} else {
		// 删除值
		err = key.DeleteValue("UDT")
	}
	
	if err != nil {
		return fmt.Errorf("注册表操作失败: %v", err)
	}
	return nil
}
