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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/FurqanSoftware/pog"
	"github.com/fatih/color"
)

var (
	flagConfig     string
	flagRooms      string
	flagRoomsSep   string
	flagRoomPrefix string

	version = ""
	date    = ""

	repoOwner = "FurqanSoftware"
	repoName  = "toph-printd"

	printdID = newID()
)

func main() {
	log.SetPrefix("\033[2K\r")
	log.SetFlags(log.Ldate | log.Ltime)

	pog.InitDefault()
	initPogHooks()

	printBanner()

	flag.StringVar(&flagConfig, "config", "printd-config.toml", "path to configuration file")
	flag.StringVar(&flagRooms, "rooms", "", "list of rooms, up to 50")
	flag.StringVar(&flagRoomsSep, "roomssep", ",", "separator string for rooms flag")
	flag.StringVar(&flagRoomPrefix, "roomprefix", "", "select rooms by prefix, ignored if rooms is set")
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
	if flagRooms != "" {
		cfg.Scope.Rooms = strings.Split(flagRooms, flagRoomsSep)
	}
	if flagRoomPrefix != "" {
		cfg.Scope.RoomPrefix = flagRoomPrefix
	}
	validateConfig(cfg)
	logConfigSummary(cfg)

	err = checkUpdate(ctx)
	if err != nil {
		pog.Warn("Could not check for updates")
	}

	if !color.NoColor {
		color.NoColor = !cfg.Printd.LogColor
	}

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
	abortch := make(chan error, 1)
	sigch := make(chan os.Signal, 2)

	pog.Info("Connecting to Toph")
	params, err := fetchParameters(ctx, cfg)
	var terr tophError
	if errors.As(err, &terr) {
		pog.Fatal(err)
	}
	catch(err)

	pog.Infof("∟ Contest: %s", ellipsize(params.ContestTitle, 35, "..."))

	if params.ContestLocked {
		pog.Info("Contest is locked")
		close(abortch)

	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pulseLoop(ctx, cfg, printdID, exitch)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			pog.Info("Waiting for prints")
			if len(cfg.Scope.Rooms) > 0 {
				pog.Infof("∟ Rooms: %s", strings.Join(cfg.Scope.Rooms, ", "))
			} else if cfg.Scope.RoomPrefix != "" {
				pog.Infof("∟ Rooms (Prefix): %s", cfg.Scope.RoomPrefix)
			}
			Daemon{
				cfg:           cfg,
				params:        params,
				exitCh:        exitch,
				abortCh:       abortch,
				pog:           pog.Default(),
				delayNotFound: 5 * time.Second,
			}.Loop(ctx)
		}()

		signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
		pog.Info("Press Ctrl+C to exit")
	}

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

func printBanner() {
	fmt.Fprintln(log.Writer(), `  ____       _       _      _ 
 |  _ \ _ __(_)_ __ | |_ __| |
`+" | |_) | '__| | '_ \\| __/ _` |"+`
 |  __/| |  | | | | | || (_| |
 |_|   |_|  |_|_| |_|\__\__,_|`)
	fmt.Fprintln(log.Writer())
	fmt.Fprintln(log.Writer(), "For Toph, By Furqan Software (https://furqansoftware.com)")
	fmt.Fprintln(log.Writer())

	if version != "" {
		fmt.Fprintf(log.Writer(), "» Release: %s", version)
	} else {
		fmt.Fprint(log.Writer(), "» Release: -")
	}
	if date != "" {
		fmt.Fprintf(log.Writer(), " (%s)", date)
	}
	fmt.Fprintln(log.Writer())
	fmt.Fprintln(log.Writer())

	fmt.Fprintf(log.Writer(), "» Project: https://github.com/%s/%s\n", repoOwner, repoName)
	fmt.Fprintln(log.Writer(), "» Support: https://community.toph.co/c/support/printd/57")
	fmt.Fprintln(log.Writer())
}

func catch(err error) {
	if err != nil {
		if version == "" {
			panic(err)
		} else {
			pog.Fatalln("Fatal error:", err)
		}
	}
}

func ellipsize(s string, n int, end string) string {
	r := []rune(s)
	if len(r) < n {
		return s
	}
	return strings.TrimSpace(string(r[:n])) + end
}
