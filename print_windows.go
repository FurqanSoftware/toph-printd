package main

import (
	"os/exec"
	"strconv"

	"github.com/FurqanSoftware/pog"
)

func printPDF(cfg Config, name string) error {
	args := []string{}
	args = append(args, name)
	if cfg.Printer.Name != "" {
		args = append(args, strconv.Quote(cfg.Printer.Name))
	}
	cmd := exec.Command(`.\PDFtoPrinter.exe`, args...)
	return cmd.Run()
}

func checkDependencies() {
	_, err := exec.LookPath(`.\PDFtoPrinter.exe`)
	if err != nil {
		pog.Fatal("Missing dependency: could not find PDFtoPrinter.exe")
	}
}

func checkPrinter(cfg Config) error {
	// TODO: We have to implement printer check for Windows.
	return nil
}
