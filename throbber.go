package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Throbber struct {
	state ThrobberState
	m     sync.Mutex
}

func (t *Throbber) SetState(s ThrobberState) {
	t.m.Lock()
	t.state = s
	t.m.Unlock()
}

func (t *Throbber) Loop(cfg Config, exitch chan struct{}) {
	pad := strings.Repeat(" ", 20)
L:
	for i := 0; ; i = (i + 1) % 10 {
		var s string
		t.m.Lock()
		switch t.state {
		case ThrobberReady:
			b := []byte{' '}
			if i < 5 {
				b[0] = '~'
			}
			s = color.GreenString("[" + string(b) + "]")
			s += " Ready"
		case ThrobberPrinting:
			s = color.BlueString("[~]")
			s += " Printing"
		case ThrobberOffline:
			s = color.RedString("[!]")
			s += " Offline"
		}
		t.m.Unlock()
		fmt.Fprintf(log.Writer(), "\033[2K\r%s%s\r", pad, s)

		select {
		case <-exitch:
			break L
		case <-time.After(125 * time.Millisecond):
		}
	}
}

type ThrobberState int

const (
	ThrobberReady ThrobberState = iota
	ThrobberPrinting
	ThrobberOffline
)
