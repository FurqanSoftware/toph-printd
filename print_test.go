package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNextPrintForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "test-token"
	cfg.Toph.ContestID = "abc123"

	_, err := getNextPrint(context.Background(), cfg)
	assert.Error(t, err)

	var terr tophError
	assert.True(t, errors.As(err, &terr))
	assert.ErrorIs(t, err, errInvalidToken)
	assert.False(t, errors.As(err, &retryableError{}))
}

func TestGetNextPrintServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "test-token"
	cfg.Toph.ContestID = "abc123"

	_, err := getNextPrint(context.Background(), cfg)
	assert.Error(t, err)

	// 503 response body is not valid JSON, so it should fail at decode and be retryable.
	assert.True(t, errors.As(err, &retryableError{}))
}

func TestGetNextPrintContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should never reach here.
		t.Fatal("request should not have been made")
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "test-token"
	cfg.Toph.ContestID = "abc123"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := getNextPrint(ctx, cfg)
	assert.Error(t, err)
	assert.True(t, errors.As(err, &retryableError{}))
}

func TestGetNextPrintNotFoundContestLocked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Toph-Contest-Locked", "1")
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "test-token"
	cfg.Toph.ContestID = "abc123"

	_, err := getNextPrint(context.Background(), cfg)
	assert.Error(t, err)

	var perr noNextPrintError
	assert.True(t, errors.As(err, &perr))
	assert.True(t, perr.contestLocked)
}

func TestGetNextPrintNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "test-token"
	cfg.Toph.ContestID = "abc123"

	_, err := getNextPrint(context.Background(), cfg)
	assert.Error(t, err)

	var perr noNextPrintError
	assert.True(t, errors.As(err, &perr))
	assert.False(t, perr.contestLocked)
}
