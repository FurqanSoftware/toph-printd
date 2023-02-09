package main

import (
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
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
	FontSize     int
	LineHeight   float64
	MarginTop    float64
	MarginRight  float64
	MarginBottom float64
	MarginLeft   float64
	TabSize      int
	KeepPDF      bool
	DelayAfter   time.Duration
	DelayError   time.Duration
	LogColor     bool
	Throbber     bool
}

func (c *ConfigPrintd) initDefaults() {
	c.FontSize = 13
	c.LineHeight = 20
	c.MarginTop = 50
	c.MarginRight = 25
	c.MarginBottom = 50
	c.MarginLeft = 25
	c.TabSize = 4
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
		log.Fatalln(color.RedString("[E]"), "Incomplete configuration: missing Toph base URL")
	}
	if cfg.Toph.Token == "" {
		log.Fatalln(color.RedString("[E]"), "Incomplete configuration: missing token")
	}
	if cfg.Toph.ContestID == "" {
		log.Fatalln(color.RedString("[E]"), "Incomplete configuration: missing contest ID")
	}
}

type PageSize string

const (
	PageA4     PageSize = "A4"
	PageLetter PageSize = "letter"
	PageLegal  PageSize = "legal"
)
