package main

import (
	"fmt"
	"os"

	"github.com/FurqanSoftware/pog"
)

func initPogHooks() {
	pog.AddExitHook(func() {
		if os.Getenv("PROMPT") != "" {
			return
		}

		pog.Info("Press Enter to close the window")
		fmt.Scanln()
	})
}
