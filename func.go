package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	Timestamp = "timestamp"
	Skip      = "skip"
	Overwrite = "overwrite"
)

// 扫描目录下所有文件，并拷贝匹配正则的文件
func scanAndCopyFiles(sourceDir, targetDir string, regexes []*regexp.Regexp, maxSizeMB int, conflictStrategy string) error {
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
			match := regex.MatchString(unixPath)
			if match {
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

				buf := bufio.NewReaderSize(sourceFile, 2*1024*1024) // 2MB缓冲区
				_, err = io.CopyBuffer(targetFile, buf, make([]byte, 2*1024*1024))
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

// 获取 Windows 临时目录
func getTempDir() string {
	return os.TempDir()
}

func safeFileName(label string) string {
	// 定义Windows系统文件名非法字符
	invalidChars := `\/:*?"<>|`

	// 替换非法字符为下划线
	result := strings.Map(func(r rune) rune {
		if strings.ContainsRune(invalidChars, r) {
			return '_'
		}
		return r
	}, label)

	// 去除首尾空格和点（Windows文件名结尾不允许空格和点）
	result = strings.TrimRight(result, " .")
	if result == "" {
		return "Untitled" // 空文件名保护
	}
	if isReservedName(result) {
		result += "_"
	}
	return result
}

func isReservedName(name string) bool {
	reserved := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	upper := strings.ToUpper(name)
	for _, r := range reserved {
		if upper == r {
			return true
		}
	}
	return false
}
