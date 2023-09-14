package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Parameters struct {
	ContestTitle  string
	ContestLocked bool
}

func fetchParameters(ctx context.Context, cfg Config) (params Parameters, err error) {
	b := bytes.Buffer{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/printd/contests/%s/parameters", cfg.Toph.BaseURL, cfg.Toph.ContestID), nil)
	if err != nil {
		return Parameters{}, err
	}
	req.Header.Add("Authorization", "Printd "+cfg.Toph.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Parameters{}, retryableError{tophError{"Could not reach Toph", err}}
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusNotFound:
		return Parameters{}, nil
	case http.StatusForbidden:
		return Parameters{}, tophError{"Could not retrieve parameters", errInvalidToken}
	}

	b.Reset()
	_, err = io.Copy(&b, resp.Body)
	if err != nil {
		return Parameters{}, retryableError{tophError{"Could not retrieve parameters", err}}
	}

	err = json.NewDecoder(&b).Decode(&params)
	if err != nil {
		return Parameters{}, retryableError{tophError{"Could not parse response", err}}
	}
	return params, nil
}
