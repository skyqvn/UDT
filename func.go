package main

import (
	"github.com/dlclark/regexp2"
	"golang.org/x/sys/windows"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	Timestamp = "timestamp"
	Skip      = "skip"
	Overwrite = "overwrite"
)

// 扫描目录下所有文件，并拷贝匹配正则的文件
func scanAndCopyFiles(sourceDir, targetDir string, regexPatterns []string, maxSizeMB int, conflictStrategy string) error {
	var regexes []*regexp2.Regexp
	for _, pattern := range regexPatterns {
		regex, err := regexp2.Compile(pattern, 0)
		if err != nil {
			return err
		}
		regexes = append(regexes, regex)
	}
	
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error scanning files: %v", err)
			return nil
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			log.Printf("Error scanning files: %v", err)
			return nil
		}
		// 统一转换为UNIX路径格式
		unixPath := strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		
		// 对文件名进行正则匹配，只要满足其中一个正则表达式即可
		for _, regex := range regexes {
			match, _ := regex.MatchString(unixPath)
			if match {
				// 检查文件大小
				fileSizeMB := info.Size() / (1024 * 1024)
				if maxSizeMB != -1 && fileSizeMB > int64(maxSizeMB) {
					log.Printf("跳过文件 %s，因为其大小 %d MB 超过了限制 %d MB", path, fileSizeMB, maxSizeMB)
					return nil
				}
				
				sourceFile, err := os.Open(path)
				if err != nil {
					log.Printf("Error scanning and copying files: %v", err)
					return nil
				}
				defer sourceFile.Close()
				
				relativePath, err := filepath.Rel(sourceDir, path)
				if err != nil {
					log.Printf("Error scanning and copying files: %v", err)
					return nil
				}
				targetFilePath := filepath.Join(targetDir, relativePath)
				if destInfo, err := os.Stat(targetFilePath); err == nil {
					switch conflictStrategy {
					case Timestamp:
						if info.ModTime().Before(destInfo.ModTime()) {
							log.Printf("跳过较旧文件: %s", targetFilePath)
							return nil
						}
					case Skip:
						log.Printf("跳过已存在文件: %s", targetFilePath)
						return nil
					case Overwrite:
						// 直接继续执行覆盖
					}
				}
				tempFilePath := targetFilePath + ".part"
				targetDirPath := filepath.Dir(tempFilePath)
				if _, err := os.Stat(targetDirPath); os.IsNotExist(err) {
					err = os.MkdirAll(targetDirPath, os.ModePerm)
					if err != nil {
						log.Printf("Error scanning and copying files: %v", err)
						return nil
					}
				}
				
				targetFile, err := os.Create(tempFilePath)
				if err != nil {
					log.Printf("Error scanning and copying files: %v", err)
					return nil
				}
				
				_, err = io.Copy(targetFile, sourceFile)
				if err != nil {
					targetFile.Close()
					os.Remove(tempFilePath) // 拷贝失败时删除临时文件
					log.Printf("Error scanning and copying files: %v", err)
					return nil
				}
				err = targetFile.Close()
				if err != nil {
					os.Remove(tempFilePath) // 拷贝失败时删除临时文件
					return nil
				}
				if err == nil {
					// 强制覆盖重命名
					if _, err := os.Stat(targetFilePath); err == nil {
						os.Remove(targetFilePath) // 先删除已存在文件
					}
					if renameErr := os.Rename(tempFilePath, targetFilePath); renameErr != nil {
						log.Printf("重命名文件失败: %v", renameErr)
					} else {
						log.Printf("成功拷贝文件: %s", targetFilePath)
					}
				}
				break // 只要匹配一个正则表达式就进行拷贝，然后跳出循环
			}
		}
		
		return nil
	})
}

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
	return os.TempDir()
}
