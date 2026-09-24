//go:build windows && !dev

package main

import (
	_ "embed"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

//go:embed build/windows/icon.ico
var appIconICO []byte

var (
	user32                            = syscall.NewLazyDLL("user32.dll")
	kernel32                          = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleWindow              = kernel32.NewProc("GetConsoleWindow")
	procFreeConsole                   = kernel32.NewProc("FreeConsole")
	procGetCurrentProcessId           = kernel32.NewProc("GetCurrentProcessId")
	procSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	procSetProcessDPIAware            = user32.NewProc("SetProcessDPIAware")
	procLookupIconIdFromDirectoryEx   = user32.NewProc("LookupIconIdFromDirectoryEx")
	procCreateIconFromResourceEx      = user32.NewProc("CreateIconFromResourceEx")
	procSendMessageW                  = user32.NewProc("SendMessageW")
	procEnumWindows                   = user32.NewProc("EnumWindows")
	procGetWindowThreadProcessId      = user32.NewProc("GetWindowThreadProcessId")
	procIsWindowVisible               = user32.NewProc("IsWindowVisible")
	procGetSystemMetrics              = user32.NewProc("GetSystemMetrics")
)

const (
	createNewProcessGroup           = 0x00000200
	detachedProcess                 = 0x00000008
	dpiAwarenessContextPerMonitorV2 = 0xFFFFFFFE
	smCXIcon                        = 11
	smCYIcon                        = 12
	smCXSmIcon                      = 49
	smCYSmIcon                      = 50
	wmSetIcon                       = 0x0080
	iconBig                         = 1
	iconSmall                       = 0
	lrDefaultSize                   = 0x00000040
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

func setWindowIcon() {
	if len(appIconICO) == 0 {
		return
	}
	hwnd := findMainWindow()
	if hwnd == 0 {
		return
	}
	base := uintptr(unsafe.Pointer(&appIconICO[0]))
	for _, spec := range []struct{ cx, cy, smcx, smcy int32 }{
		{0, 0, smCXIcon, smCYIcon},
		{0, 0, smCXSmIcon, smCYSmIcon},
	} {
		cx, _, _ := procGetSystemMetrics.Call(uintptr(spec.smcx))
		cy, _, _ := procGetSystemMetrics.Call(uintptr(spec.smcy))
		offset, _, _ := procLookupIconIdFromDirectoryEx.Call(base, 1, cx, cy, lrDefaultSize)
		if offset == 0 {
			continue
		}
		hicon, _, _ := procCreateIconFromResourceEx.Call(
			base+offset,
			uintptr(len(appIconICO))-offset,
			1, 0x00030000, cx, cy, lrDefaultSize,
		)
		if hicon == 0 {
			continue
		}
		which := uintptr(iconBig)
		if spec.smcx == smCXSmIcon {
			which = iconSmall
		}
		procSendMessageW.Call(hwnd, wmSetIcon, which, hicon)
	}
}

func findMainWindow() uintptr {
	pid, _, _ := procGetCurrentProcessId.Call()
	var found uintptr
	cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		var winPid uint32
		procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&winPid)))
		if uintptr(winPid) != pid {
			return 1
		}
		if v, _, _ := procIsWindowVisible.Call(hwnd); v == 0 {
			return 1
		}
		found = hwnd
		return 0
	})
	procEnumWindows.Call(cb, 0)
	return found
}
