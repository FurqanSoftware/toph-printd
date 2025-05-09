package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/FurqanSoftware/pog"
	"github.com/alexbrainman/printer"
)

func printPDF(cfg Config, name string) error {
	args := []string{}
	args = append(args, name)
	if cfg.Printer.Name != "" {
		args = append(args, strconv.Quote(cfg.Printer.Name))
	}
	cmd := exec.Command(`.\PDFtoPrinter.exe`, args...)
	_, err := cmd.Output()
	if err != nil {
		var exiterr *exec.ExitError
		if errors.As(err, &exiterr) {
			return printDispatchError{fmt.Errorf("%w: %s", err, ellipsize(string(exiterr.Stderr), 50, "..."))}
		}
		return printDispatchError{err}
	}
	return nil
}

func checkDependencies() {
	_, err := exec.LookPath(`.\PDFtoPrinter.exe`)
	if err != nil {
		pog.Fatal("Missing dependency: could not find PDFtoPrinter.exe")
	}
}

func checkPrinter(cfg Config) error {
	names, err := printer.ReadNames()
	if err != nil {
		return err
	}
	for _, name := range names {
		if cfg.Printer.Name == "" {
			return nil
		}
		if name == cfg.Printer.Name {
			return nil
		}
	}
	return errPrinterNotExist
}
