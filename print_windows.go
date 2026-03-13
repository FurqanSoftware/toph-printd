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
	var cmd *exec.Cmd
	switch cfg.Windows.PrintHelper {
	case PrintHelperSumatraPDF:
		args := []string{"-print-to-default", "-silent", name}
		if cfg.Printer.Name != "" {
			args = []string{"-print-to", cfg.Printer.Name, "-silent", name}
		}
		cmd = exec.Command(`.\SumatraPDF.exe`, args...)
	default:
		args := []string{name}
		if cfg.Printer.Name != "" {
			args = append(args, strconv.Quote(cfg.Printer.Name))
		}
		cmd = exec.Command(`.\PDFtoPrinter.exe`, args...)
	}
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

func resolvePrintHelper(cfg *Config) {
	if cfg.Windows.PrintHelper != PrintHelperAuto {
		return
	}
	if _, err := exec.LookPath(`.\PDFtoPrinter.exe`); err == nil {
		cfg.Windows.PrintHelper = PrintHelperPDFtoPrinter
		return
	}
	if _, err := exec.LookPath(`.\SumatraPDF.exe`); err == nil {
		cfg.Windows.PrintHelper = PrintHelperSumatraPDF
		return
	}
	pog.Fatal("Missing dependency: could not find PDFtoPrinter.exe or SumatraPDF.exe")
}

func checkDependencies(cfg Config) {
	switch cfg.Windows.PrintHelper {
	case PrintHelperSumatraPDF:
		_, err := exec.LookPath(`.\SumatraPDF.exe`)
		if err != nil {
			pog.Fatal("Missing dependency: could not find SumatraPDF.exe")
		}
	default:
		_, err := exec.LookPath(`.\PDFtoPrinter.exe`)
		if err != nil {
			pog.Fatal("Missing dependency: could not find PDFtoPrinter.exe")
		}
	}
}

var printHelperNames = map[PrintHelper]string{
	PrintHelperPDFtoPrinter: "PDFtoPrinter",
	PrintHelperSumatraPDF:   "SumatraPDF",
}

func logPlatformConfigSummary(cfg Config) {
	name := printHelperNames[cfg.Windows.PrintHelper]
	if name == "" {
		name = string(cfg.Windows.PrintHelper)
	}
	pog.Infof("∟ Print Helper: %s", name)
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
