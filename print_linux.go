package main

import (
	"bufio"
	"bytes"
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

func checkPrinter(cfg Config) error {
	cmd := exec.Command("lpstat", "-p")
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Split(bufio.ScanLines)
	for sc.Scan() {
		line := sc.Bytes()
		fields := bytes.Fields(line)
		if len(fields) < 2 || !bytes.Equal(fields[0], []byte("printer")) {
			continue
		}
		if cfg.Printer.Name == "" {
			return nil
		}
		if bytes.Equal(fields[1], []byte(cfg.Printer.Name)) {
			return nil
		}
	}
	return errPrinterNotExist
}
