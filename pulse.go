package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func pulseLoop(cfg Config) {
	for {
		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/printd/pulse?contest=%s", cfg.Toph.BaseURL, cfg.Toph.ContestID), nil)
		req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
		if err != nil {
			log.Println(err)
			continue
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}
		resp.Body.Close()

		time.Sleep(5 * time.Second)
	}
}
