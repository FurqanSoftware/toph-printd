package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/FurqanSoftware/pog"
)

func initPogHooks() {
	pog.AddExitHook(func() {
		if !startedByExplorer() {
			return
		}

		pog.Info("Press Enter to close the window")
		fmt.Scanln()
	})
}

func startedByExplorer() bool {
	ppid := syscall.Getppid()

	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(snapshot)
	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	if err = syscall.Process32First(snapshot, &entry); err != nil {
		return false
	}
	for {
		if entry.ProcessID == uint32(ppid) {
			return syscall.UTF16ToString(entry.ExeFile[:]) == "explorer.exe"
		}
		err = syscall.Process32Next(snapshot, &entry)
		if err != nil {
			return false
		}
	}
}
