package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Print struct {
	ID         string
	Header     string
	Content    string
	Status     string
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func getNextPrint(ctx context.Context, cfg Config) (Print, error) {
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/printd/contests/%s/next_print", cfg.Toph.BaseURL, cfg.Contest.ID), nil)
		req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
		if err != nil {
			return Print{}, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return Print{}, err
		}
		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			time.Sleep(5 * time.Second)
			continue
		} else {
			defer resp.Body.Close()
		}

		pr := Print{}
		err = json.NewDecoder(resp.Body).Decode(&pr)
		return pr, err
	}
}

func runPrintJob(ctx context.Context, cfg Config, pr Print) error {
	name := pr.ID + ".pdf"
	err := PDFBuilder{
		cfg: cfg,
	}.Build(name, pr)
	if err != nil {
		return err
	}

	err = printPDF(cfg, name)
	if err != nil {
		return err
	}

	if !cfg.Printd.KeepPDF {
		err = os.Remove(name)
		if err != nil {
			return err
		}
	}

	return nil
}

func markPrintDone(ctx context.Context, cfg Config, pr Print) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/printd/prints/%s/mark_done?contest=%s", cfg.Toph.BaseURL, pr.ID, cfg.Contest.ID), nil)
	req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}