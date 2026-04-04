package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/FurqanSoftware/pog"
)

func pulseLoop(ctx context.Context, cfg Config, printdid string, exitch chan struct{}) {
L:
	for {
		var resp *http.Response
		req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/printd/pulse?contest=%s", cfg.Toph.BaseURL, cfg.Toph.ContestID), nil)
		if err != nil {
			pog.Error(err)
			goto Retry
		}
		req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
		req.Header.Add("Printd-ID", printdid)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			pog.Errorln("Could not send pulse:", err)
			goto Retry
		}
		resp.Body.Close()

	Retry:
		select {
		case <-exitch:
			break L
		case <-time.After(5 * time.Second):
		}
	}
}
