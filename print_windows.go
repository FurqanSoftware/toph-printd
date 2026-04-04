package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/FurqanSoftware/pog"
	"github.com/alexbrainman/printer"
)

func printPDF(cfg Config, name string) error {
	var cmd *exec.Cmd
	switch cfg.Windows.PrintHelper {
	case PrintHelperSumatraPDF:
		args := []string{"-print-to-default", "-silent", "-exit-when-done", name}
		if cfg.Printer.Name != "" {
			args = []string{"-print-to", cfg.Printer.Name, "-silent", "-exit-when-done", name}
		}
		cmd = exec.Command(cfg.Windows.PrintHelperPath, args...)
	default:
		args := []string{name}
		if cfg.Printer.Name != "" {
			args = append(args, strconv.Quote(cfg.Printer.Name))
		}
		cmd = exec.Command(cfg.Windows.PrintHelperPath, args...)
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

var printHelperExe = map[PrintHelper]string{
	PrintHelperPDFtoPrinter: `.\PDFtoPrinter.exe`,
	PrintHelperSumatraPDF:   `.\SumatraPDF.exe`,
}

func resolvePrintHelper(cfg *Config) {
	if cfg.Windows.PrintHelper == PrintHelperAuto {
		for _, h := range []PrintHelper{PrintHelperPDFtoPrinter, PrintHelperSumatraPDF} {
			if _, err := exec.LookPath(printHelperExe[h]); err == nil {
				cfg.Windows.PrintHelper = h
				goto PinPath
			}
		}
		pog.Fatal("Missing dependency: could not find PDFtoPrinter.exe or SumatraPDF.exe")
		return
	}

PinPath:
	exe, ok := printHelperExe[cfg.Windows.PrintHelper]
	if !ok {
		pog.Fatal("Invalid print helper: " + string(cfg.Windows.PrintHelper))
	}
	p, err := exec.LookPath(exe)
	if err != nil {
		pog.Fatal("Missing dependency: could not find " + exe)
	}
	cfg.Windows.PrintHelperPath, _ = filepath.Abs(p)
}

func checkDependencies(cfg Config) {}

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
