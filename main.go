package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/FurqanSoftware/pog"
	"github.com/avast/retry-go"
	"github.com/fatih/color"
)

var (
	flagConfig string

	version = ""
	commit  = ""
	date    = ""
	builtBy = ""

	repoOwner = "FurqanSoftware"
	repoName  = "toph-printd"
)

func main() {
	log.SetPrefix("\033[2K\r")
	log.SetFlags(log.Ldate | log.Ltime)

	pog.InitDefault()

	fmt.Fprintln(log.Writer(), `  ____       _       _      _ 
 |  _ \ _ __(_)_ __ | |_ __| |
`+" | |_) | '__| | '_ \\| __/ _` |"+`
 |  __/| |  | | | | | || (_| |
 |_|   |_|  |_|_| |_|\__\__,_|`)
	fmt.Fprintln(log.Writer())
	fmt.Fprintln(log.Writer(), "For Toph, By Furqan Software (https://furqansoftware.com)")
	fmt.Fprintln(log.Writer())

	if version == "" {
		version = "devel"
	}
	fmt.Fprintf(log.Writer(), "» Release: %s", version)
	if date != "" {
		fmt.Fprintf(log.Writer(), " (%s)", date)
	}
	fmt.Fprintln(log.Writer())
	fmt.Fprintln(log.Writer())

	fmt.Fprintf(log.Writer(), "» Project: https://github.com/%s/%s\n", repoOwner, repoName)
	fmt.Fprintln(log.Writer(), "» Support: https://community.toph.co/c/support/printd/57")
	fmt.Fprintln(log.Writer())

	flag.StringVar(&flagConfig, "config", "printd-config.toml", "path to configuration file")
	flag.Parse()

	ctx := context.Background()

	pog.Info("Loading configuration")
	cfg, err := parseConfig()
	switch {
	case errors.Is(err, fs.ErrNotExist):
		pog.Fatalf("Configuration file %s does not exist", flagConfig)
	case err != nil:
		pog.Fatal("Could not parse configuration file")
	}
	catch(err)
	validateConfig(cfg)

	if cfg.Printer.Name == "" {
		pog.Info("∟ Printer: ‹System Default›")
	} else {
		pog.Infof("∟ Printer: %s", cfg.Printer.Name)
	}
	pog.Infof("∟ Page Size: %s", cfg.Printer.PageSize)

	err = checkUpdate(ctx)
	if err != nil {
		pog.Warn("Could not check for updates")
	}

	color.NoColor = !cfg.Printd.LogColor

	checkDependencies()
	err = checkPrinter(cfg)
	if errors.Is(err, errPrinterNotExist) {
		if cfg.Printer.Name != "" {
			pog.Fatalf("Printer %s does not exist", cfg.Printer.Name)
		} else {
			pog.Fatal("No printer exists")
		}
	}

	http.DefaultClient.Timeout = cfg.Toph.Timeout

	wg := sync.WaitGroup{}
	exitch := make(chan struct{})
	abortch := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		pulseLoop(cfg, exitch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		pog.Info("Waiting for prints")
		delay := 0 * time.Second
	L:
		for {
			pr, err := getNextPrint(ctx, cfg)
			var terr tophError
			if errors.As(err, &terr) {
				pog.SetStatus(statusOffline)
				pog.Error(err)
				if !errors.As(err, &retryableError{}) {
					close(abortch)
					break L
				}
				delay = cfg.Printd.DelayError
				goto retry
			}
			catch(err)

			if pr.ID == "" {
				pog.SetStatus(statusReady)
				delay = 5 * time.Second
				goto retry
			}

			pog.SetStatus(statusPrinting)

			pog.Infof("Printing %s", pr.ID)
			err = runPrintJob(ctx, cfg, pr)
			catch(err)
			err = retry.Do(func() error {
				return markPrintDone(ctx, cfg, pr)
			},
				retry.RetryIf(func(err error) bool { return errors.As(err, &retryableError{}) }),
				retry.Attempts(3),
				retry.Delay(500*time.Millisecond),
				retry.LastErrorOnly(true),
			)
			if errors.As(err, &terr) {
				pog.SetStatus(statusOffline)
				pog.Error(err)
				if !errors.As(err, &retryableError{}) {
					close(abortch)
					break L
				}
				delay = cfg.Printd.DelayError
				goto retry
			}
			catch(err)
			pog.Info("∟ Done")

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
		pog.Infof("Received %s", sig)
	case <-abortch:
	}

	pog.Info("Exiting")
	pog.Stop()
	close(exitch)
	wg.Wait()

	pog.Info("Goodbye")
}

func catch(err error) {
	if err != nil {
		if version == "devel" {
			panic(err)
		} else {
			pog.Fatalln("Fatal error:", err)
		}
	}
}
