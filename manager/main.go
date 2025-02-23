package main

import (
	"atomicgo.dev/cursor"
	_ "embed"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 1 {
		execute(os.Args[1])
	} else {
		err := enableANSI()
		cursor.Hide()
		if err != nil {
			fmt.Println(err, "\r")
		}
		// 定义按钮列表
		buttons := []Button{
			{
				Text: "Install",
				Action: func() {
					start()
					err := runAsAdmin(os.Args[0], "install")
					if err != nil {
						fmt.Printf("以管理员权限启动程序失败: %v\r\n", err)
					}
					os.Exit(0)
				},
			},
			{
				Text: "Uninstall",
				Action: func() {
					err := runAsAdmin(os.Args[0], "uninstall")
					if err != nil {
						fmt.Printf("以管理员权限启动程序失败: %v\r\n", err)
					}
					os.Exit(0)
				},
			},
			{
				Text: "Stop",
				Action: func() {
					stop()
					pause()
					os.Exit(0)
				},
			},
			{
				Text: "Restart",
				Action: func() {
					reboot()
					pause()
					os.Exit(0)
				},
			},
			{
				Text: "Edit",
				Action: func() {
					edit()
					pause()
					os.Exit(0)
				},
			},
		}
		
		// 创建菜单
		menu := NewMenu(buttons)
		
		// 运行菜单
		menu.Run()
	}
}
