package main

import (
	"log"
	"os/exec"
	"strconv"
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
		log.Fatalln("Missing dependency: could not find PDFtoPrinter.exe")
	}
}
