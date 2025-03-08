package main

import "syscall"

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	GenerateConsoleCtrlEvent = kernel32.NewProc("GenerateConsoleCtrlEvent")
	SetConsoleCtrlHandler    = kernel32.NewProc("SetConsoleCtrlHandler")
	AttachConsole            = kernel32.NewProc("AttachConsole")
	FreeConsole              = kernel32.NewProc("FreeConsole")
	shell32                  = syscall.NewLazyDLL("shell32.dll")
	ShellExecuteW            = shell32.NewProc("ShellExecuteW")
)
