package main

import (
	"log"
	"os/exec"
	"strconv"

	"github.com/fatih/color"
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
		log.Fatalln(color.RedString("[E]"), "Missing dependency: could not find PDFtoPrinter.exe")
	}
}

func checkPrinter(cfg Config) error {
	// TODO: We have to implement printer check for Windows.
	return nil
}
