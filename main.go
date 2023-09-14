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

	fmt.Fprintln(log.Writer(), `  ____       _       _      _ 
 |  _ \ _ __(_)_ __ | |_ __| |
`+" | |_) | '__| | '_ \\| __/ _` |"+`
 |  __/| |  | | | | | || (_| |
 |_|   |_|  |_|_| |_|\__\__,_|`)
	fmt.Fprintln(log.Writer())
	fmt.Fprintln(log.Writer(), "For Toph, By Furqan Software (https://furqansoftware.com)")
	fmt.Fprintln(log.Writer())

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

	log.Println("[i]", "Loading configuration")
	cfg, err := parseConfig()
	switch {
	case errors.Is(err, fs.ErrNotExist):
		log.Fatalln(color.RedString("[E]"), fmt.Sprintf("Configuration file %s does not exist", flagConfig))
	case err != nil:
		log.Fatalln(color.RedString("[E]"), fmt.Sprintf("Could not parse configuration file"))
	}
	catch(err)
	validateConfig(cfg)

	err = checkUpdate(ctx)
	if err != nil {
		log.Println(color.HiYellowString("[W]"), "Could not check for updates")
	}

	color.NoColor = !cfg.Printd.LogColor

	checkDependencies()
	err = checkPrinter(cfg)
	if errors.Is(err, errPrinterNotExist) {
		if cfg.Printer.Name != "" {
			log.Fatalln(color.RedString("[E]"), fmt.Sprintf("Printer %s does not exist", cfg.Printer.Name))
		} else {
			log.Fatalln(color.RedString("[E]"), "No printer exists")
		}
	}

	http.DefaultClient.Timeout = cfg.Toph.Timeout

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
			var terr tophError
			if errors.As(err, &terr) {
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

	sig := <-sigch
	log.Printf("[i]"+" Received %s", sig)

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
