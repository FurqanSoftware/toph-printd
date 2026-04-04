package main

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/FurqanSoftware/pog"
)

type Config struct {
	Printd  ConfigPrintd
	Printer ConfigPrinter
	Toph    ConfigToph
	Scope   ConfigScope
	Windows ConfigWindows
	Debug   ConfigDebug
}

func (c *Config) initDefaults() {
	c.Printd.initDefaults()
	c.Printer.initDefaults()
	c.Toph.initDefaults()
	c.Windows.initDefaults()
	c.Debug.initDefaults()
}

type ConfigPrintd struct {
	FontSize         int
	LineHeight       float64
	MarginTop        float64
	MarginRight      float64
	MarginBottom     float64
	MarginLeft       float64
	TabSize          int
	HeaderExtra      string
	ReduceBlankLines bool
	KeepPDF          bool
	DelayAfter       time.Duration
	DelayError       time.Duration
	LogColor         bool
	Throbber         bool
}

func (c *ConfigPrintd) initDefaults() {
	c.FontSize = 13
	c.LineHeight = 20
	c.MarginTop = 50
	c.MarginRight = 25
	c.MarginBottom = 50
	c.MarginLeft = 25
	c.TabSize = 4
	c.HeaderExtra = ""
	c.ReduceBlankLines = false
	c.KeepPDF = true
	c.DelayAfter = 500 * time.Millisecond
	c.DelayError = 5 * time.Second
	c.LogColor = true
	c.Throbber = true
}

type PrintHelper string

const (
	PrintHelperAuto         PrintHelper = "auto"
	PrintHelperPDFtoPrinter PrintHelper = "pdf-to-printer"
	PrintHelperSumatraPDF   PrintHelper = "sumatra-pdf"
)

type ConfigPrinter struct {
	Name     string
	PageSize PageSize
}

func (c *ConfigPrinter) initDefaults() {
	c.PageSize = PageA4
}

type ConfigWindows struct {
	PrintHelper     PrintHelper
	PrintHelperPath string `toml:"-"`
}

func (c *ConfigWindows) initDefaults() {
	c.PrintHelper = PrintHelperAuto
}

type ConfigToph struct {
	BaseURL   string
	Token     string
	ContestID string
	Timeout   time.Duration
}

func (c *ConfigToph) initDefaults() {
	c.BaseURL = "https://toph.co"
	c.Timeout = 30 * time.Second
}

type ConfigScope struct {
	Rooms      []string
	RoomPrefix string
}

type ConfigDebug struct {
	DontPrint bool
}

func (c *ConfigDebug) initDefaults() {
	c.DontPrint = false
}

func parseConfig() (cfg Config, err error) {
	cfg.initDefaults()

	b, err := os.ReadFile(flagConfig)
	if err != nil {
		return
	}
	_, err = toml.Decode(string(b), &cfg)
	if err != nil {
		return
	}
	return
}

func validateConfig(cfg Config) error {
	var msg []string
	if cfg.Toph.BaseURL == "" {
		msg = append(msg, "missing Toph base URL")
	}
	if cfg.Toph.Token == "" {
		msg = append(msg, "missing token")
	}
	if cfg.Toph.ContestID == "" {
		msg = append(msg, "missing contest ID")
	}
	if len(msg) != 0 {
		return errors.New(strings.Join(msg, ", "))
	}
	return nil
}

func logConfigSummary(cfg Config) {
	if cfg.Printer.Name == "" {
		pog.Info("∟ Printer: ‹System Default›")
	} else {
		pog.Infof("∟ Printer: %s", cfg.Printer.Name)
	}
	pog.Infof("∟ Page Size: %s", cfg.Printer.PageSize)
	logPlatformConfigSummary(cfg)
}

type PageSize string

const (
	PageA4     PageSize = "A4"
	PageLetter PageSize = "letter"
	PageLegal  PageSize = "legal"
)
