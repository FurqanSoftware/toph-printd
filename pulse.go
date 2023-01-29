package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func pulseLoop(cfg Config, exitch chan struct{}) {
L:
	for {
		var resp *http.Response
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/printd/pulse?contest=%s", cfg.Toph.BaseURL, cfg.Toph.ContestID), nil)
		req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
		if err != nil {
			log.Println(err)
			goto retry
		}
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			goto retry
		}
		resp.Body.Close()

	retry:
		select {
		case <-exitch:
			break L
		case <-time.After(5 * time.Second):
		}
	}
}
