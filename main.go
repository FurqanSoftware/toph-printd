package main

import (
	"context"
	"log"
)

func main() {
	log.SetFlags(0)

	ctx := context.Background()

	log.Println("Loading configuration")
	cfg, err := parseConfig()
	catch(err)
	log.Printf(".. %#v", cfg)

	go pulseLoop(cfg)

	for {
		log.Println("Waiting for prints")
		pr, err := getNextPrint(ctx, cfg)
		catch(err)

		log.Printf("Printing %s", pr.ID)
		err = runPrintJob(ctx, cfg, pr)
		catch(err)
		err = markPrintDone(ctx, cfg, pr)
		catch(err)
		log.Printf(".. Done")
	}
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}