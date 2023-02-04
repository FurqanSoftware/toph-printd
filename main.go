package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
)

var (
	flagConfig string
)

func main() {
	log.SetPrefix("\033[2K\r")
	log.SetFlags(log.Ldate | log.Ltime)

	fmt.Fprintln(log.Writer(), `  ____       _       _      _ 
 |  _ \ _ __(_)_ __ | |_ __| |
`+" | |_) | '__| | '_ \\| __/ _` |"+`
 |  __/| |  | | | | | || (_| |
 |_|   |_|  |_|_| |_|\__\__,_|
`)
	fmt.Fprintln(log.Writer(), "For Toph, By Furqan Software (https://furqansoftware.com)")
	fmt.Fprintln(log.Writer())

	flag.StringVar(&flagConfig, "config", "printd-config.toml", "path to configuration file")
	flag.Parse()

	ctx := context.Background()

	checkDependencies()

	log.Println("[i]", "Loading configuration")
	cfg, err := parseConfig()
	catch(err)
	validateConfig(cfg)

	color.NoColor = !cfg.Printd.LogColor

	wg := sync.WaitGroup{}
	exitch := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		pulseLoop(cfg, exitch)
	}()

	throbber := Throbber{}
	if cfg.Printd.Throbber {
		go throbber.Loop(cfg, exitch)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("[i]", "Waiting for prints")
		delay := 0 * time.Second
	L:
		for {
			pr, err := getNextPrint(ctx, cfg)
			if errors.As(err, &tophError{}) {
				throbber.SetState(ThrobberOffline)
				log.Println(color.RedString("[E]"), err)
				delay = cfg.Printd.DelayError
				goto retry
			}
			catch(err)

			if pr.ID == "" {
				throbber.SetState(ThrobberReady)
				delay = 5 * time.Second
				goto retry
			}

			throbber.SetState(ThrobberPrinting)

			log.Printf("[i]"+" Printing %s", pr.ID)
			err = runPrintJob(ctx, cfg, pr)
			catch(err)
			err = markPrintDone(ctx, cfg, pr)
			catch(err)
			log.Println("[i]", ".. Done")

			delay = cfg.Printd.DelayAfter

		retry:
			select {
			case <-exitch:
				break L
			case <-time.After(delay):
			}
		}
	}()

	sigch := make(chan os.Signal, 2)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigch:
		log.Printf("[i]"+" Received %s", sig)
	}

	log.Println("[i]", "Exiting")
	close(exitch)
	wg.Wait()

	log.Println("[i]", "Goodbye")
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
