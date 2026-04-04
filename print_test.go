package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
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
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"

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
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"

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
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"

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
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"

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
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"

	_, err := getNextPrint(context.Background(), cfg)
	assert.Error(t, err)

	var perr noNextPrintError
	assert.True(t, errors.As(err, &perr))
	assert.False(t, perr.contestLocked)
}

func TestGetNextPrintRoomFiltering(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rooms := r.URL.Query()["rooms"]
		assert.Equal(t, []string{"A", "B"}, rooms)
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"
	cfg.Scope.Rooms = []string{"A", "B"}

	getNextPrint(context.Background(), cfg)
}

func TestGetNextPrintRoomPrefix(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "LAB", r.URL.Query().Get("roomprefix"))
		assert.Empty(t, r.URL.Query()["rooms"])
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"
	cfg.Scope.RoomPrefix = "LAB"

	getNextPrint(context.Background(), cfg)
}

func TestGetNextPrintAuthorizationHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Printd keyboardcat", r.Header.Get("Authorization"))
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"

	getNextPrint(context.Background(), cfg)
}

func TestMarkPrintDonePayload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Printd keyboardcat", r.Header.Get("Authorization"))
		assert.Equal(t, "6502f105025832238e865526", r.URL.Query().Get("contest"))

		var done Done
		err := json.NewDecoder(r.Body).Decode(&done)
		assert.NoError(t, err)
		assert.Equal(t, 3, done.PageCount)
		assert.Equal(t, 1, done.PageSkipped)

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = "keyboardcat"
	cfg.Toph.ContestID = "6502f105025832238e865526"

	pr := Print{ID: "6502f46b17592a5a9e870928"}
	pdf := PDF{PageCount: 3, PageSkipped: 1}

	err := markPrintDone(context.Background(), cfg, pr, pdf)
	assert.NoError(t, err)
}

func TestRunPrintJobTempFile(t *testing.T) {
	cfg := Config{}
	cfg.initDefaults()
	cfg.Printd.KeepPDF = false
	cfg.Debug.DontPrint = true

	pr := Print{
		ID:        "6502f46b17592a5a9e870928",
		Header:    "Test Print",
		Content:   "Hello, World!",
		PageLimit: -1,
	}

	pdf, err := runPrintJob(context.Background(), cfg, pr)
	assert.NoError(t, err)
	assert.Equal(t, 1, pdf.PageCount)

	// PDF should not exist in CWD when KeepPDF is false.
	assert.NoFileExists(t, "6502f46b17592a5a9e870928.pdf")
}

func TestRunPrintJobKeepPDF(t *testing.T) {
	cfg := Config{}
	cfg.initDefaults()
	cfg.Printd.KeepPDF = true
	cfg.Debug.DontPrint = true

	pr := Print{
		ID:        "6502f9bf92c75c2f7874698e",
		Header:    "Test Print",
		Content:   "Hello, World!",
		PageLimit: -1,
	}

	pdf, err := runPrintJob(context.Background(), cfg, pr)
	assert.NoError(t, err)
	assert.Equal(t, 1, pdf.PageCount)

	assert.FileExists(t, "6502f9bf92c75c2f7874698e.pdf")
	os.Remove("6502f9bf92c75c2f7874698e.pdf")
}
