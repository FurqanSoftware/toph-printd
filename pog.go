package main

import (
	"github.com/fatih/color"
)

type pogStatus struct {
	icon  byte
	text  string
	color *color.Color
	throb bool
}

func (s pogStatus) Icon() byte          { return s.icon }
func (s pogStatus) Text() string        { return s.text }
func (s pogStatus) Color() *color.Color { return s.color }
func (s pogStatus) Throb() bool         { return s.throb }

var (
	statusReady    = pogStatus{'~', "Ready", color.New(color.FgGreen), true}
	statusPrinting = pogStatus{'~', "Printing", color.New(color.FgBlue), false}
	statusOffline  = pogStatus{'!', "Offline", color.New(color.FgRed), false}
)
