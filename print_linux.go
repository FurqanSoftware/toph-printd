package main

import (
	"log"
	"os/exec"

	"github.com/fatih/color"
)

func printPDF(cfg Config, name string) error {
	args := []string{}
	if cfg.Printer.Name != "" {
		args = append(args, "-P", cfg.Printer.Name)
	}
	args = append(args, name)
	cmd := exec.Command("lpr", args...)
	return cmd.Run()
}

func checkDependencies() {
	_, err := exec.LookPath("lpr")
	if err != nil {
		log.Fatalln(color.RedString("[E]"), "Missing dependency: could not find lpr")
	}
}
