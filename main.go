package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	
	"github.com/spf13/viper"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows"
)

var query = "SELECT DeviceID,DriveType,VolumeName FROM Win32_LogicalDisk WHERE DriveType=2" // DriveType=2表示可移动设备

type win32LogicalDisk struct {
	DeviceID   string
	DriveType  uint32
	VolumeName string
}

func init() {
	// 获取可执行文件的路径
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("获取可执行文件位置错误: %v", err)
	}
	// 获取可执行文件所在的目录
	exeDir := filepath.Dir(exePath)
	// 更改当前工作目录为可执行文件所在的目录
	err = os.Chdir(exeDir)
	if err != nil {
		log.Fatalf("改变工作目录错误: %v", err)
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
	
	var regexes []*regexp.Regexp
	for _, pattern := range regexPatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			log.Fatalf("正则表达式编译错误：%v", err)
		}
		regexes = append(regexes, regex)
	}
	
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
	signal.Notify(sigs,
		syscall.Signal(windows.CTRL_C_EVENT),
		syscall.Signal(windows.CTRL_BREAK_EVENT),
		syscall.Signal(windows.CTRL_CLOSE_EVENT),
		syscall.Signal(windows.CTRL_LOGOFF_EVENT),
		syscall.Signal(windows.CTRL_SHUTDOWN_EVENT),
		windows.SIGTERM,
	)
	go func() {
		<-sigs
		log.Println("收到退出信号。正在释放锁并退出...")
		f.Close()
		os.Remove(lockFile)
		os.Exit(0)
	}()
	
	// 用于存储已挂载的 U 盘路径
	usbDrives := make(map[string]bool, 26)
	
	for {
		usbList, err := getUSBDrives()
		if err == nil {
		loop:
			for drive, label := range usbList {
				if !usbDrives[drive] {
					for _, excluded := range excludeLabels {
						if label == excluded {
							log.Printf("跳过排除卷标的设备 盘符: %s 卷标: %s", drive, label)
							continue loop
						}
					}
					log.Printf("检测到U盘插入 盘符: %s 卷标: %s", drive, label)
					target := filepath.Join(targetDir, safeFileName(label))
					go scanAndCopyFiles(drive+`\`, target, regexes, maxSizeMB, conflictStrategy)
					usbDrives[drive] = true
				}
			}
			
			// 处理移除设备
			for drive := range usbDrives {
				if _, ok := usbList[drive]; !ok {
					log.Printf("U盘移除 盘符: %s", drive)
					delete(usbDrives, drive)
				}
			}
		}
		
		time.Sleep(3 * time.Second)
	}
}

// 替换原有设备检测逻辑
func getUSBDrives() (map[string]string, error) {
	var disks []win32LogicalDisk
	err := wmi.Query(query, &disks)
	
	result := make(map[string]string)
	for _, d := range disks {
		result[d.DeviceID] = d.VolumeName
	}
	return result, err
}
