package main

import "os/exec"

func printPDF(cfg Config, name string) error {
	args := []string{}
	if cfg.Printer.Name != "" {
		args = append(args, "-P", cfg.Printer.Name)
	}
	args = append(args, name)
	cmd := exec.Command("lpr", args...)
	return cmd.Run()
}
