package main

import (
	"golang.org/x/sys/windows"
)

func init() {
	enableColors()
}

func enableColors() {
	handle, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		return
	}

	var mode uint32
	err = windows.GetConsoleMode(handle, &mode)
	if err != nil {
		return
	}
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING {
		return
	}

	windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}
