package main

import "golang.org/x/sys/windows"

// 修改后的USB设备检测函数
func isUsbDrive(driveLetter string) bool {
	rootPath := driveLetter + `\`
	driveType := windows.GetDriveType(windows.StringToUTF16Ptr(rootPath))
	return driveType == windows.DRIVE_REMOVABLE
}

// 新增获取卷标函数
func getVolumeLabel(driveLetter string) string {
	rootPath := driveLetter + `\`
	buf := make([]uint16, 256)
	err := windows.GetVolumeInformation(
		windows.StringToUTF16Ptr(rootPath),
		&buf[0],
		uint32(len(buf)),
		nil,
		nil,
		nil,
		nil,
		0)
	
	if err != nil {
		return "Unknown"
	}
	return windows.UTF16ToString(buf)
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
