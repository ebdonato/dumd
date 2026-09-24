//go:build windows

package main

import "syscall"

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procFreeConsole = kernel32.NewProc("FreeConsole")
)

func init() {
	procFreeConsole.Call()
}
