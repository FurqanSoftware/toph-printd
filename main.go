package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
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

	log.Println("Loading configuration")
	cfg, err := parseConfig()
	catch(err)
	validateConfig(cfg)

	go pulseLoop(cfg)

	for {
		log.Println("Waiting for prints")
		pr, err := getNextPrint(ctx, cfg)
		if errors.As(err, &tophError{}) {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		catch(err)

		log.Printf("Printing %s", pr.ID)
		err = runPrintJob(ctx, cfg, pr)
		catch(err)
		err = markPrintDone(ctx, cfg, pr)
		catch(err)
		log.Printf(".. Done")

		time.Sleep(cfg.Printd.DelayAfter)
	}
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
