package main

import (
	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	_ "embed"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
)

//go:embed logo.txt
var logo string

// Button 定义按钮结构体
type Button struct {
	Text   string
	Action func()
}

// Menu 定义菜单结构体
type Menu struct {
	Buttons       []Button
	SelectedIndex int
}

// NewMenu 创建一个新的菜单
func NewMenu(buttons []Button) *Menu {
	return &Menu{
		Buttons:       buttons,
		SelectedIndex: 0,
	}
}

// Run 运行菜单，监听键盘事件
func (m *Menu) Run() {
	m.Draw()
	
	err := keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Up:
			if m.SelectedIndex > 0 {
				m.SelectedIndex--
			}
		case keys.Down:
			if m.SelectedIndex < len(m.Buttons)-1 {
				m.SelectedIndex++
			}
		case keys.Enter:
			m.Buttons[m.SelectedIndex].Action()
		case keys.Escape:
			return true, nil // 按下 ESC 键退出程序
		}
		m.Draw()
		return false, nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "监听键盘事件时出错:", err, "\r")
		os.Exit(1)
	}
}

func enableANSI() error {
	handle, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return err
	}
	var mode uint32
	err = windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return err
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	return windows.SetConsoleMode(handle, mode)
}

func showTitle() {
	fmt.Print("\x1b[1;32mUDT管理器\x1b[0m\r\n")
}

func printLogo() {
	fmt.Print(logo, "\r\n\r\n")
}

func showShortcutHelp() {
	fmt.Print("\r\n\n\n\n\x1b[90m")
	fmt.Print("↑/↓ - 选择按钮, Enter - 执行操作, ESC - 退出\r\n")
	fmt.Print("\x1b[0m")
}

// Draw 绘制菜单界面
func (m *Menu) Draw() {
	fmt.Print("\x1b[2J\x1b[H") // 清屏
	showTitle()
	printLogo()
	fmt.Print("\x1b[1m功能: \x1b[0m\r\n")
	for i, btn := range m.Buttons {
		if i == m.SelectedIndex {
			fmt.Printf("> \x1b[7m%s\x1b[0m\r\n", btn.Text)
		} else {
			fmt.Printf("  %s\r\n", btn.Text)
		}
	}
	showShortcutHelp()
	fmt.Print("\x1b[0m")
}
