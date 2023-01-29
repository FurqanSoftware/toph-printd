package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

func getNextPrint(ctx context.Context, cfg Config) (pr Print, err error) {
	b := bytes.Buffer{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/printd/contests/%s/next_print", cfg.Toph.BaseURL, cfg.Toph.ContestID), nil)
	if err != nil {
		return Print{}, err
	}
	req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Print{}, tophError{"Could not reach Toph", err}
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return Print{}, nil
	}

	b.Reset()
	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		return Print{}, tophError{"Could not retrieve print", err}
	}
	resp.Body.Close()

	err = json.NewDecoder(&b).Decode(&pr)
	if err != nil {
		return Print{}, tophError{"Could not parse response", err}
	}
	return pr, nil
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
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/printd/prints/%s/mark_done?contest=%s", cfg.Toph.BaseURL, pr.ID, cfg.Toph.ContestID), nil)
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
