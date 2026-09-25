//go:build windows && !dev

package main

import (
	"os"
	"os/exec"
	"syscall"
)

var (
	user32                            = syscall.NewLazyDLL("user32.dll")
	kernel32                          = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleWindow              = kernel32.NewProc("GetConsoleWindow")
	procFreeConsole                   = kernel32.NewProc("FreeConsole")
	procSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	procSetProcessDPIAware            = user32.NewProc("SetProcessDPIAware")
)

const (
	createNewProcessGroup           = 0x00000200
	detachedProcess                 = 0x00000008
	dpiAwarenessContextPerMonitorV2 = 0xFFFFFFFE
)

func prepareWindows() {
	if r, _, _ := procSetProcessDpiAwarenessContext.Call(dpiAwarenessContextPerMonitorV2); r == 0 {
		procSetProcessDPIAware.Call()
	}

	if os.Getenv("DUMD_DETACHED") == "1" {
		return
	}
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		procFreeConsole.Call()
		return
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Env = append(os.Environ(), "DUMD_DETACHED=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup | detachedProcess}
	if err := cmd.Start(); err != nil {
		procFreeConsole.Call()
		return
	}
	os.Exit(0)
}
