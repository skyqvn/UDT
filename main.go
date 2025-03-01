package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	
	"github.com/spf13/viper"
	"golang.org/x/sys/windows"
)

func init() {
	// 获取可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}
	// 获取可执行文件所在的目录
	exeDir := filepath.Dir(exePath)
	// 更改当前工作目录为可执行文件所在的目录
	err = os.Chdir(exeDir)
	if err != nil {
		log.Fatalf("Error changing working directory: %v", err)
	}
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
		log.Println("服务已禁用")
		return
	}
	maxSizeMB := viper.GetInt("maxSizeMB")
	targetDir := viper.GetString("targetDir")
	regexPatterns := viper.GetStringSlice("regexPatterns")
	conflictStrategy := viper.GetString("conflictStrategy")
	excludeLabels := viper.GetStringSlice("excludeVolumeLabels")
	
	// 构建 UDT 目录路径
	udtDir := filepath.Join(getTempDir(), "UDT")
	// 检查 UDT 目录是否存在，如果不存在则创建它
	if _, err := os.Stat(udtDir); os.IsNotExist(err) {
		err := os.MkdirAll(udtDir, os.ModePerm)
		if err != nil {
			log.Fatalf("创建UDT目录时出错：%v", err)
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
				log.Fatalf("读取锁文件错误：%v", err)
			}
			pidStr := strings.TrimSpace(string(content))
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				log.Fatalf("从锁文件中解析进程ID时出错：%v", err)
			}
			log.Printf("应用程序的另一个实例已经在运行，其进程ID为：%d。正在退出…", pid)
			os.Exit(0)
		}
		log.Fatalf("创建锁文件错误：%v", err)
	}
	// 将当前进程号写入锁文件
	pid := os.Getpid()
	_, err = f.WriteString(strconv.Itoa(pid))
	if err != nil {
		log.Fatalf("将进程ID写入锁文件时出错：%v", err)
	}
	defer func() {
		f.Close()
		os.Remove(lockFile)
	}()
	
	// 注册信号处理函数，在收到退出信号时释放锁
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.Signal(10), syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigs
		log.Println("收到退出信号。正在释放锁并退出...")
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
			log.Fatalf("获取逻辑驱动器时出错：%v", err)
		}
	
	mainloop:
		for i := 0; i < 26; i++ {
			if bitmask&(1<<i) != 0 {
				driveLetter := string('A'+i) + ":"
				if isUsbDrive(driveLetter) {
					if _, ok := usbDrives[driveLetter]; !ok {
						usbDrives[driveLetter] = true
						label := getVolumeLabel(driveLetter)
						// 卷标排除检查
						for _, excluded := range excludeLabels {
							if label == excluded {
								log.Printf("跳过排除卷标的设备 盘符: %s 卷标: %s", driveLetter, label)
								continue mainloop
							}
						}
						log.Printf("检测到U盘插入 盘符: %s 卷标: %s", driveLetter, label)
						target := filepath.Join(targetDir, label)
						scanAndCopyFiles(driveLetter+`\`, target, regexPatterns, maxSizeMB, conflictStrategy)
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
