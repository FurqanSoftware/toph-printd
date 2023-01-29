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
)

var (
	flagConfig string
)

func main() {
	fmt.Println(`  ____       _       _      _ 
 |  _ \ _ __(_)_ __ | |_ __| |
` + " | |_) | '__| | '_ \\| __/ _` |" + `
 |  __/| |  | | | | | || (_| |
 |_|   |_|  |_|_| |_|\__\__,_|
`)
	fmt.Println("For Toph, By Furqan Software (https://furqansoftware.com)")
	fmt.Println()

	flag.StringVar(&flagConfig, "config", "printd-config.toml", "path to configuration file")
	flag.Parse()

	ctx := context.Background()

	checkDependencies()

	log.Println("Loading configuration")
	cfg, err := parseConfig()
	catch(err)
	validateConfig(cfg)

	wg := sync.WaitGroup{}
	exitch := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		pulseLoop(cfg, exitch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		delay := 0 * time.Second
	L:
		for {
			log.Println("Waiting for prints")
			pr, err := getNextPrint(ctx, cfg)
			if errors.As(err, &tophError{}) {
				log.Println(err)
				delay = cfg.Printd.DelayError
				goto retry
			}
			catch(err)

			log.Printf("Printing %s", pr.ID)
			err = runPrintJob(ctx, cfg, pr)
			catch(err)
			err = markPrintDone(ctx, cfg, pr)
			catch(err)
			log.Printf(".. Done")

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
		log.Printf("Received %s", sig)
	}

	log.Println("Exiting")
	close(exitch)
	wg.Wait()

	log.Println("Goodbye")
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
