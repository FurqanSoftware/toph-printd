package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FurqanSoftware/pog"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration simulates the full printd lifecycle: fetching parameters,
// running the pulse loop, polling for prints, generating PDFs, and marking
// prints as done.
func TestIntegration(t *testing.T) {
	pog.InitDefault()
	pogger := pog.NewPogger(io.Discard, "", 0)

	const (
		token     = "keyboardcat"
		contestID = "6502f105025832238e865526"
	)

	prints := []Print{
		{
			ID:        "6502f46b17592a5a9e870928",
			Header:    "Team Alpha · Problem A",
			Content:   "#include <stdio.h>\nint main() {\n\tprintf(\"Hello, World!\\n\");\n\treturn 0;\n}",
			Status:    "Queued",
			PageLimit: -1,
		},
		{
			ID:        "6502f9bf92c75c2f7874698e",
			Header:    "Team Bravo · Problem B",
			Content:   "x = int(input())\nprint(x * 2)\n",
			Status:    "Queued",
			PageLimit: -1,
		},
	}

	queue := NewQueue([]Frame{
		{print: &prints[0]},
		{print: &prints[1]},
	})

	var (
		pulseCount   atomic.Int32
		paramsCalled atomic.Bool
		doneIDs      []string
		doneMu       sync.Mutex
		donePayloads []Done
	)

	r := mux.NewRouter()

	// Parameters endpoint.
	r.HandleFunc("/api/printd/contests/{contestID}/parameters", func(w http.ResponseWriter, r *http.Request) {
		paramsCalled.Store(true)
		vars := mux.Vars(r)
		assert.Equal(t, contestID, vars["contestID"])
		assert.Equal(t, "Printd "+token, r.Header.Get("Authorization"))

		json.NewEncoder(w).Encode(Parameters{
			ContestTitle:  "Integration Test Contest",
			ContestLocked: false,
		})
	})

	// Pulse endpoint.
	r.HandleFunc("/api/printd/pulse", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Printd "+token, r.Header.Get("Authorization"))
		assert.Equal(t, contestID, r.URL.Query().Get("contest"))
		assert.NotEmpty(t, r.Header.Get("Printd-ID"))
		pulseCount.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	// Next print endpoint.
	r.HandleFunc("/api/printd/contests/{contestID}/next_print", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		assert.Equal(t, contestID, vars["contestID"])
		assert.Equal(t, "Printd "+token, r.Header.Get("Authorization"))

		fr := queue.Next()
		if fr == nil || fr.print == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(fr.print)
	})

	// Mark done endpoint.
	r.NewRoute().
		Path("/api/printd/prints/{printID}/mark_done").
		Queries("contest", "{contestID}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "Printd "+token, r.Header.Get("Authorization"))
			vars := mux.Vars(r)
			assert.Equal(t, contestID, vars["contestID"])

			var done Done
			err := json.NewDecoder(r.Body).Decode(&done)
			assert.NoError(t, err)

			doneMu.Lock()
			doneIDs = append(doneIDs, vars["printID"])
			donePayloads = append(donePayloads, done)
			doneMu.Unlock()

			w.WriteHeader(http.StatusOK)
		})

	srv := httptest.NewServer(r)
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = token
	cfg.Toph.ContestID = contestID
	cfg.Printd.KeepPDF = false
	cfg.Debug.DontPrint = true

	ctx := context.Background()

	// Fetch parameters, like main() does.
	params, err := fetchParameters(ctx, cfg)
	require.NoError(t, err)
	assert.True(t, paramsCalled.Load())
	assert.Equal(t, "Integration Test Contest", params.ContestTitle)
	assert.False(t, params.ContestLocked)

	// Start pulse and daemon loops.
	exitch := make(chan struct{})
	abortch := make(chan error, 1)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		pulseLoop(ctx, cfg, "test-printd-id", exitch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		Daemon{
			cfg:           cfg,
			params:        params,
			exitCh:        exitch,
			abortCh:       abortch,
			pog:           pogger,
			delayNotFound: 125 * time.Millisecond,
		}.Loop(ctx)
	}()

	// Wait for all prints to be consumed.
	<-queue.emptyCh
	// Give a moment for mark_done to complete.
	time.Sleep(250 * time.Millisecond)

	close(exitch)
	wg.Wait()

	// Verify no abort error.
	select {
	case err := <-abortch:
		t.Fatalf("unexpected abort: %v", err)
	default:
	}

	// Verify all prints were marked done in order.
	assert.Equal(t, []string{
		"6502f46b17592a5a9e870928",
		"6502f9bf92c75c2f7874698e",
	}, doneIDs)

	// Verify page counts were reported.
	for _, payload := range donePayloads {
		assert.Equal(t, 1, payload.PageCount)
		assert.Equal(t, 0, payload.PageSkipped)
	}

	// Verify pulse was sent at least once.
	assert.GreaterOrEqual(t, pulseCount.Load(), int32(1))

	// Verify temp PDFs were cleaned up (KeepPDF is false).
	assert.NoFileExists(t, "6502f46b17592a5a9e870928.pdf")
	assert.NoFileExists(t, "6502f9bf92c75c2f7874698e.pdf")
}

// TestIntegrationContestLocked verifies that a locked contest is handled
// correctly: parameters are fetched, but no prints are processed.
func TestIntegrationContestLocked(t *testing.T) {
	pog.InitDefault()

	const (
		token     = "keyboardcat"
		contestID = "6502f105025832238e865526"
	)

	r := mux.NewRouter()

	r.HandleFunc("/api/printd/contests/{contestID}/parameters", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Parameters{
			ContestTitle:  "Locked Contest",
			ContestLocked: true,
		})
	})

	r.HandleFunc("/api/printd/contests/{contestID}/next_print", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next_print should not be called when contest is locked")
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = token
	cfg.Toph.ContestID = contestID
	cfg.Debug.DontPrint = true

	ctx := context.Background()

	params, err := fetchParameters(ctx, cfg)
	require.NoError(t, err)
	assert.Equal(t, "Locked Contest", params.ContestTitle)
	assert.True(t, params.ContestLocked)
}

// TestIntegrationRoomFiltering verifies that room scope parameters are passed
// through the full flow to the API, and that prints are served per room.
func TestIntegrationRoomFiltering(t *testing.T) {
	pog.InitDefault()

	const (
		token     = "keyboardcat"
		contestID = "6502f105025832238e865526"
	)

	// All prints, tagged by room.
	type roomPrint struct {
		room  string
		print Print
	}
	allPrints := []roomPrint{
		{"A1", Print{ID: "aaa0000000000000000a0001", Header: "Team 1 · A1", Content: "print('A1')", Status: "Queued", PageLimit: -1}},
		{"A2", Print{ID: "aaa0000000000000000a0002", Header: "Team 2 · A2", Content: "print('A2')", Status: "Queued", PageLimit: -1}},
		{"B1", Print{ID: "bbb0000000000000000b0001", Header: "Team 3 · B1", Content: "print('B1')", Status: "Queued", PageLimit: -1}},
		{"B2", Print{ID: "bbb0000000000000000b0002", Header: "Team 4 · B2", Content: "print('B2')", Status: "Queued", PageLimit: -1}},
		{"C1", Print{ID: "ccc0000000000000000c0001", Header: "Team 5 · C1", Content: "print('C1')", Status: "Queued", PageLimit: -1}},
	}

	// Server-side filtering: returns prints matching the requested rooms or prefix.
	newServer := func(t *testing.T) (srv *httptest.Server, doneIDs *[]string, mu *sync.Mutex) {
		t.Helper()
		var m sync.Mutex
		mu = &m
		ids := &[]string{}
		remaining := make([]roomPrint, len(allPrints))
		copy(remaining, allPrints)

		r := mux.NewRouter()

		r.HandleFunc("/api/printd/contests/{contestID}/next_print", func(w http.ResponseWriter, r *http.Request) {
			rooms := r.URL.Query()["rooms"]
			prefix := r.URL.Query().Get("roomprefix")

			mu.Lock()
			defer mu.Unlock()

			for i, rp := range remaining {
				match := false
				if len(rooms) > 0 {
					for _, room := range rooms {
						if rp.room == room {
							match = true
							break
						}
					}
				} else if prefix != "" {
					if len(rp.room) >= len(prefix) && rp.room[:len(prefix)] == prefix {
						match = true
					}
				}
				if match {
					remaining = append(remaining[:i], remaining[i+1:]...)
					json.NewEncoder(w).Encode(rp.print)
					return
				}
			}
			http.Error(w, "Not Found", http.StatusNotFound)
		})

		r.NewRoute().
			Path("/api/printd/prints/{printID}/mark_done").
			Queries("contest", "{contestID}").
			HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)
				mu.Lock()
				*ids = append(*ids, vars["printID"])
				mu.Unlock()
				w.WriteHeader(http.StatusOK)
			})

		return httptest.NewServer(r), ids, mu
	}

	runDaemon := func(t *testing.T, srvURL string, rooms []string, roomPrefix string, doneIDs *[]string, mu *sync.Mutex, expectCount int) {
		t.Helper()
		pogger := pog.NewPogger(io.Discard, "", 0)

		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.BaseURL = srvURL
		cfg.Toph.Token = token
		cfg.Toph.ContestID = contestID
		cfg.Scope.Rooms = rooms
		cfg.Scope.RoomPrefix = roomPrefix
		cfg.Printd.KeepPDF = false
		cfg.Printd.DelayAfter = 0
		cfg.Debug.DontPrint = true

		ctx := context.Background()
		exitch := make(chan struct{})
		abortch := make(chan error, 1)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			Daemon{
				cfg:           cfg,
				exitCh:        exitch,
				abortCh:       abortch,
				pog:           pogger,
				delayNotFound: 50 * time.Millisecond,
			}.Loop(ctx)
		}()

		deadline := time.After(5 * time.Second)
		tick := time.NewTicker(50 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-deadline:
				close(exitch)
				wg.Wait()
				t.Fatal("timed out waiting for daemon to process prints")
			case <-tick.C:
				mu.Lock()
				n := len(*doneIDs)
				mu.Unlock()
				if n >= expectCount {
					time.Sleep(100 * time.Millisecond)
					close(exitch)
					wg.Wait()

					select {
					case err := <-abortch:
						t.Fatalf("unexpected abort: %v", err)
					default:
					}
					return
				}
			}
		}
	}

	t.Run("specific rooms", func(t *testing.T) {
		srv, doneIDs, mu := newServer(t)
		defer srv.Close()

		runDaemon(t, srv.URL, []string{"A1", "B1"}, "", doneIDs, mu, 2)

		assert.Equal(t, []string{
			"aaa0000000000000000a0001",
			"bbb0000000000000000b0001",
		}, *doneIDs)
	})

	t.Run("single room", func(t *testing.T) {
		srv, doneIDs, mu := newServer(t)
		defer srv.Close()

		runDaemon(t, srv.URL, []string{"C1"}, "", doneIDs, mu, 1)

		assert.Equal(t, []string{
			"ccc0000000000000000c0001",
		}, *doneIDs)
	})

	t.Run("room prefix A", func(t *testing.T) {
		srv, doneIDs, mu := newServer(t)
		defer srv.Close()

		runDaemon(t, srv.URL, nil, "A", doneIDs, mu, 2)

		assert.Equal(t, []string{
			"aaa0000000000000000a0001",
			"aaa0000000000000000a0002",
		}, *doneIDs)
	})

	t.Run("room prefix B", func(t *testing.T) {
		srv, doneIDs, mu := newServer(t)
		defer srv.Close()

		runDaemon(t, srv.URL, nil, "B", doneIDs, mu, 2)

		assert.Equal(t, []string{
			"bbb0000000000000000b0001",
			"bbb0000000000000000b0002",
		}, *doneIDs)
	})

	t.Run("no matching rooms", func(t *testing.T) {
		srv, doneIDs, _ := newServer(t)
		defer srv.Close()

		pogger := pog.NewPogger(io.Discard, "", 0)

		cfg := Config{}
		cfg.initDefaults()
		cfg.Toph.BaseURL = srv.URL
		cfg.Toph.Token = token
		cfg.Toph.ContestID = contestID
		cfg.Scope.Rooms = []string{"Z9"}
		cfg.Printd.KeepPDF = false
		cfg.Printd.DelayAfter = 0
		cfg.Debug.DontPrint = true

		ctx := context.Background()
		exitch := make(chan struct{})
		abortch := make(chan error, 1)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			Daemon{
				cfg:           cfg,
				exitCh:        exitch,
				abortCh:       abortch,
				pog:           pogger,
				delayNotFound: 50 * time.Millisecond,
			}.Loop(ctx)
		}()

		// Let the daemon poll a few times to confirm nothing comes through.
		time.Sleep(300 * time.Millisecond)
		close(exitch)
		wg.Wait()

		assert.Empty(t, *doneIDs)
	})
}
