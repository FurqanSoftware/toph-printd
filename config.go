package main

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/FurqanSoftware/pog"
)

type Config struct {
	Printd  ConfigPrintd
	Printer ConfigPrinter
	Toph    ConfigToph
}

func (c *Config) initDefaults() {
	c.Printd.initDefaults()
	c.Printer.initDefaults()
	c.Toph.initDefaults()
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

type ConfigPrinter struct {
	Name     string
	PageSize PageSize
}

func (c *ConfigPrinter) initDefaults() {
	c.PageSize = "A4"
}

type ConfigToph struct {
	BaseURL   string
	Token     string
	ContestID string
	Timeout   time.Duration
}

func (c *ConfigToph) initDefaults() {
	c.BaseURL = "https://toph.co"
	c.Timeout = 10 * time.Second
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

func validateConfig(cfg Config) {
	if cfg.Toph.BaseURL == "" {
		pog.Error("Incomplete configuration: missing Toph base URL")
	}
	if cfg.Toph.Token == "" {
		pog.Error("Incomplete configuration: missing token")
	}
	if cfg.Toph.ContestID == "" {
		pog.Error("Incomplete configuration: missing contest ID")
	}
}

func logConfigSummary(cfg Config) {
	if cfg.Printer.Name == "" {
		pog.Info("∟ Printer: ‹System Default›")
	} else {
		pog.Infof("∟ Printer: %s", cfg.Printer.Name)
	}
	pog.Infof("∟ Page Size: %s", cfg.Printer.PageSize)
}

type PageSize string

const (
	PageA4     PageSize = "A4"
	PageLetter PageSize = "letter"
	PageLegal  PageSize = "legal"
)
