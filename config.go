package main

import (
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Printd struct {
		FontSize     int
		LineHeight   float64
		MarginTop    float64
		MarginRight  float64
		MarginBottom float64
		MarginLeft   float64
		TabSize      int
		KeepPDF      bool
		DelayAfter   time.Duration
	}
	Printer struct {
		Name     string
		PageSize PageSize
	}
	Contest struct {
		ID string
	}
	Toph struct {
		BaseURL string
		Token   string
	}
}

func parseConfig() (cfg Config, err error) {
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

type PageSize string

const (
	PageA4     PageSize = "A4"
	PageLetter PageSize = "letter"
	PageLegal  PageSize = "legal"
)
