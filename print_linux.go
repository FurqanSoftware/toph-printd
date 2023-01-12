package main

import "os/exec"

func printPDF(cfg Config, name string) error {
	cmd := exec.Command("lpr", "-P", cfg.Printer.Name, name)
	return cmd.Run()
}
