package main

import (
	"atomicgo.dev/cursor"
	"fmt"
	"os"
)

const (
	InstallCurrentUser   = "install-current-user"
	InstallAllUsers      = "install-all-user"
	UninstallCurrentUser = "uninstall-current-user"
	UninstallAllUsers    = "uninstall-all-user"
)

func main() {
	enableANSI()
	if len(os.Args) != 1 {
		execute(os.Args[1])
	} else {
		cursor.Hide()
		// 定义按钮列表
		buttons := []Button{
			{
				Text: "安装（当前用户）",
				Action: func() {
					start()
					if isAdmin() {
						execute(InstallCurrentUser)
					} else {
						err := runAsAdmin(os.Args[0], InstallCurrentUser)
						if err != nil {
							fmt.Printf("以管理员权限启动程序失败: %v\r\n", err)
						}
						os.Exit(0)
					}
				},
			},
			{
				Text: "安装（所有用户）",
				Action: func() {
					start()
					if isAdmin() {
						execute(InstallAllUsers)
					} else {
						err := runAsAdmin(os.Args[0], InstallAllUsers)
						if err != nil {
							fmt.Printf("以管理员权限启动程序失败: %v\r\n", err)
						}
						os.Exit(0)
					}
				},
			},
			{
				Text: "卸载（当前用户）",
				Action: func() {
					if isAdmin() {
						execute(UninstallCurrentUser)
					} else {
						err := runAsAdmin(os.Args[0], UninstallCurrentUser)
						if err != nil {
							fmt.Printf("以管理员权限启动程序失败: %v\r\n", err)
						}
						os.Exit(0)
					}
				},
			},
			{
				Text: "卸载（所有用户）",
				Action: func() {
					if isAdmin() {
						execute(UninstallAllUsers)
					} else {
						err := runAsAdmin(os.Args[0], UninstallAllUsers)
						if err != nil {
							fmt.Printf("以管理员权限启动程序失败: %v\r\n", err)
						}
						os.Exit(0)
					}
				},
			},
			{
				Text: "停止",
				Action: func() {
					stop()
					pause()
					os.Exit(0)
				},
			},
			{
				Text: "重启",
				Action: func() {
					reboot()
					pause()
					os.Exit(0)
				},
			},
			// {
			// 	Text: "编辑配置",
			// 	Action: func() {
			// 		edit()
			// 		pause()
			// 		os.Exit(0)
			// 	},
			// },
		}
		
		// 创建菜单
		menu := NewMenu(buttons)
		
		// 运行菜单
		menu.Run()
	}
}
