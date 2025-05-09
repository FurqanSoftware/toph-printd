package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Print struct {
	ID         string
	Header     string
	Content    string
	Status     string
	PageLimit  int
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func getNextPrint(ctx context.Context, cfg Config) (pr Print, err error) {
	b := bytes.Buffer{}

	q := url.Values{}
	if len(cfg.Scope.Rooms) > 0 {
		for _, room := range cfg.Scope.Rooms {
			q.Add("rooms", room)
		}
	} else if cfg.Scope.RoomPrefix != "" {
		q.Set("roomprefix", cfg.Scope.RoomPrefix)
	}

	u := fmt.Sprintf("%s/api/printd/contests/%s/next_print", cfg.Toph.BaseURL, cfg.Toph.ContestID)
	if len(q) > 0 {
		u += "?" + q.Encode()
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return Print{}, err
	}
	req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Print{}, retryableError{tophError{"Could not reach Toph", err}}
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotFound:
		return Print{}, noNextPrintError{
			contestLocked: resp.Header.Get("Toph-Contest-Locked") == "1",
		}
	case http.StatusForbidden:
		return Print{}, tophError{"Could not retrieve print", errInvalidToken}
	}

	b.Reset()
	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		return Print{}, retryableError{tophError{"Could not retrieve print", err}}
	}

	err = json.NewDecoder(&b).Decode(&pr)
	if err != nil {
		return Print{}, retryableError{tophError{"Could not parse response", err}}
	}
	return pr, nil
}

func runPrintJob(ctx context.Context, cfg Config, pr Print) (PDF, error) {
	name := pr.ID + ".pdf"
	pdf, err := PDFBuilder{
		cfg: cfg,
	}.Build(name, pr)
	if err != nil {
		return PDF{}, err
	}

	if !cfg.Debug.DontPrint && pdf.PageCount > 0 {
		err = printPDF(cfg, name)
		if err != nil {
			return PDF{}, err
		}
	}

	if !cfg.Printd.KeepPDF {
		err = os.Remove(name)
		if err != nil {
			return PDF{}, err
		}
	}

	return pdf, nil
}

type Done struct {
	PageCount int `json:"pageCount"`
}

func markPrintDone(ctx context.Context, cfg Config, pr Print, pdf PDF) error {
	body := Done{
		PageCount: pdf.PageCount,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/printd/prints/%s/mark_done?contest=%s", cfg.Toph.BaseURL, pr.ID, cfg.Toph.ContestID), bytes.NewReader(b))
	req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return retryableError{tophError{"Could not reach Toph", err}}
	}
	defer resp.Body.Close()

	return nil
}
