package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	
	"github.com/dlclark/regexp2"
	"github.com/spf13/viper"
	"golang.org/x/sys/windows"
)

// 扫描目录下所有文件，并拷贝匹配正则的文件
func scanAndCopyFiles(sourceDir, targetDir string, regexPatterns []string, maxSizeMB int) error {
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
					log.Printf("Error scanning and copying files: %v", err)
					return nil
				}
				err = targetFile.Close()
				if err != nil {
					os.Remove(tempFilePath) // 拷贝失败时删除临时文件
				}
				if err == nil {
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

func main() {
	// 初始化 viper
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("未找到配置文件: %v", err)
		} else {
			log.Fatalf("读取配置文件时出错: %v", err)
		}
	}
	
	// 从配置文件中获取配置项
	if !viper.GetBool("enabled") {
		log.Println("服务已禁用\r")
		return
	}
	maxSizeMB := viper.GetInt("maxSizeMB")
	targetDir := viper.GetString("targetDir")
	regexPatterns := viper.GetStringSlice("regexPatterns")
	
	// 构建 UDT 目录路径
	udtDir := filepath.Join(getTempDir(), "UDT")
	// 检查 UDT 目录是否存在，如果不存在则创建它
	if _, err := os.Stat(udtDir); os.IsNotExist(err) {
		err := os.MkdirAll(udtDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Error creating UDT directory: %v", err)
		}
	}
	// 构建锁文件的完整路径
	lockFile := filepath.Join(udtDir, "app.lock")
	
	// 尝试删除锁文件，以确认是否被占用
	os.Remove(lockFile)
	f, err := os.OpenFile(lockFile, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		if os.IsExist(err) {
			// 读取锁文件中的进程号
			content, err := os.ReadFile(lockFile)
			if err != nil {
				log.Fatalf("Error reading lock file: %v", err)
			}
			pidStr := strings.TrimSpace(string(content))
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				log.Fatalf("Error parsing PID from lock file: %v", err)
			}
			log.Printf("Another instance of the application is already running with PID: %d. Exiting...", pid)
			os.Exit(0)
		}
		log.Fatalf("Error creating lock file: %v", err)
	}
	// 将当前进程号写入锁文件
	pid := os.Getpid()
	_, err = f.WriteString(strconv.Itoa(pid))
	if err != nil {
		log.Fatalf("Error writing PID to lock file: %v", err)
	}
	
	// 注册信号处理函数，在收到退出信号时释放锁
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.Signal(10), syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigs
		log.Println("Received exit signal. Releasing lock and exiting...\r")
		f.Close()
		os.Remove(lockFile)
		os.Exit(0)
	}()
	
	// 用于存储已挂载的 U 盘路径
	usbDrives := make(map[string]bool)
	
	// 持续监控设备变化
	for {
		// 枚举所有逻辑驱动器
		bitmask, err := windows.GetLogicalDrives()
		if err != nil {
			log.Fatalf("Error getting logical drives: %v", err)
		}
		
		for i := 0; i < 26; i++ {
			if bitmask&(1<<i) != 0 {
				driveLetter := string('A'+i) + ":"
				if isUsbDrive(driveLetter) {
					if _, ok := usbDrives[driveLetter]; !ok {
						usbDrives[driveLetter] = true
						label := getVolumeLabel(driveLetter)
						log.Printf("检测到U盘插入 盘符: %s 卷标: %s", driveLetter, label)
						target := filepath.Join(targetDir, label)
						scanAndCopyFiles(driveLetter+`\`, target, regexPatterns, maxSizeMB)
					}
				} else {
					if _, ok := usbDrives[driveLetter]; ok {
						log.Printf("U盘移除 盘符: %s", driveLetter)
						delete(usbDrives, driveLetter)
					}
				}
			}
		}
		
		time.Sleep(1 * time.Second)
	}
}
