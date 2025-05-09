package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/FurqanSoftware/pog"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDaemon(t *testing.T) {
	doneids, aborterr := testDaemon(t, "ZnZ0eW81bWw1NjMzc2dlYjI5azVuNThleDVzYjZ1aG8=", "6502f105025832238e865526", NewQueue([]Frame{
		{
			print: &Print{
				ID:        "6502f46b17592a5a9e870928",
				Header:    "Test Print 1",
				Content:   "Lorem ipsum dolor.",
				Status:    "Queued",
				PageLimit: -1,
			},
		},
		{
			print: &Print{
				ID:        "6502f9bf92c75c2f7874698e",
				Header:    "Test Print 2",
				Content:   "The quick brown fox...",
				Status:    "Queued",
				PageLimit: -1,
			},
		},
	}))

	assert.NoError(t, aborterr)

	assert.FileExists(t, "6502f46b17592a5a9e870928.pdf")
	os.Remove("6502f46b17592a5a9e870928.pdf")
	assert.FileExists(t, "6502f9bf92c75c2f7874698e.pdf")
	os.Remove("6502f9bf92c75c2f7874698e.pdf")

	assert.Equal(t, []string{
		"6502f46b17592a5a9e870928",
		"6502f9bf92c75c2f7874698e",
	}, doneids)
}

func TestDaemonEmptyHeader(t *testing.T) {
	doneids, aborterr := testDaemon(t, "dGlsa2c0em5lOG4waTl3MTl1d2prejVldmk5cGIycTU=", "6502fee52e058f9990cc5c6e", NewQueue([]Frame{
		{
			print: &Print{
				ID:        "6502fec79eed503a402e0b59",
				Header:    "",
				Content:   "Lorem ipsum dolor.",
				Status:    "Queued",
				PageLimit: -1,
			},
		},
	}))

	assert.NoError(t, aborterr)

	assert.FileExists(t, "6502fec79eed503a402e0b59.pdf")
	os.Remove("6502fec79eed503a402e0b59.pdf")

	assert.Equal(t, []string{
		"6502fec79eed503a402e0b59",
	}, doneids)
}

func TestDaemonEmptyContent(t *testing.T) {
	doneids, aborterr := testDaemon(t, "bWJocHBqMXA1aHRxbHpmMWo2bnNvYmhmcmVpYXlvdGQ=", "6502fee1c7bb2bf7288be9c7", NewQueue([]Frame{
		{
			print: &Print{
				ID:        "6502fecc74093fd44c060e11",
				Header:    "Keyboard Cat",
				Content:   "",
				Status:    "Queued",
				PageLimit: -1,
			},
		},
	}))

	assert.NoError(t, aborterr)

	assert.FileExists(t, "6502fecc74093fd44c060e11.pdf")
	os.Remove("6502fecc74093fd44c060e11.pdf")

	assert.Equal(t, []string{
		"6502fecc74093fd44c060e11",
	}, doneids)
}

func TestDaemonBreak(t *testing.T) {
	doneids, aborterr := testDaemon(t, "em95bHBmMTduM25wcDJyeTVucGd1bDI3MXB1d2V6ODM=", "6502fedddafbfc6ac0f20876", NewQueue([]Frame{
		{
			print: &Print{
				ID:        "6502fed2ee36d5244aade158",
				Header:    "Test Print 1",
				Content:   "Lorem ipsum dolor.",
				Status:    "Queued",
				PageLimit: -1,
			},
		},
		{},
		{
			print: &Print{
				ID:        "6502fed88ea6c9620b82bf5a",
				Header:    "Test Print 2",
				Content:   "The quick brown fox...",
				Status:    "Queued",
				PageLimit: -1,
			},
		},
	}))

	assert.NoError(t, aborterr)

	assert.FileExists(t, "6502fed2ee36d5244aade158.pdf")
	os.Remove("6502fed2ee36d5244aade158.pdf")
	assert.FileExists(t, "6502fed88ea6c9620b82bf5a.pdf")
	os.Remove("6502fed88ea6c9620b82bf5a.pdf")

	assert.Equal(t, []string{
		"6502fed2ee36d5244aade158",
		"6502fed88ea6c9620b82bf5a",
	}, doneids)
}

func TestDaemonEmpty(t *testing.T) {
	doneids, aborterr := testDaemon(t, "ZWV5aTFwc3hjNnc2c2NlZG13MHpreHUzaDc3cXhyMmg=", "6503070f91fb17000cc2e5b9", NewQueue([]Frame{}))

	assert.NoError(t, aborterr)

	assert.Equal(t, []string{}, doneids)
}

func TestDaemonLocked(t *testing.T) {
	doneids, aborterr := testDaemon(t, "bTZwamc4bzA0MXozMXRvcDFpaTVyZmh6NHFuM3phdGY=", "65035553170f1faf07aac1ad", NewQueue([]Frame{
		{
			print: &Print{
				ID:        "6503554aad610a1d499e9a70",
				Header:    "Test Print 1",
				Content:   "Lorem ipsum dolor.",
				Status:    "Queued",
				PageLimit: -1,
			},
		},
		{
			print: &Print{
				ID:        "65035545d32aebbef6dff9fe",
				Header:    "Test Print 2",
				Content:   "The quick brown fox...",
				Status:    "Queued",
				PageLimit: -1,
			},
		},
		{
			contestLocked: true,
		},
	}))

	assert.ErrorIs(t, aborterr, noNextPrintError{contestLocked: true})

	assert.FileExists(t, "6503554aad610a1d499e9a70.pdf")
	os.Remove("6503554aad610a1d499e9a70.pdf")
	assert.FileExists(t, "65035545d32aebbef6dff9fe.pdf")
	os.Remove("65035545d32aebbef6dff9fe.pdf")

	assert.Equal(t, []string{
		"6503554aad610a1d499e9a70",
		"65035545d32aebbef6dff9fe",
	}, doneids)
}

func testDaemon(t *testing.T, token, contestid string, queue *Queue) (doneids []string, aborterr error) {
	pog.InitDefault()

	r := mux.NewRouter()

	r.HandleFunc("/api/printd/contests/{contestID}/next_print", func(w http.ResponseWriter, r *http.Request) {
		assertTokenInRequest(t, r, token)
		fr := queue.Next()
		if fr == nil || fr.print == nil {
			if fr != nil && fr.contestLocked {
				w.Header().Set("Toph-Contest-Locked", "1")
			}
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		err := json.NewEncoder(w).Encode(fr.print)
		assert.NoError(t, err)
	})

	doneids = []string{}
	r.NewRoute().
		Path("/api/printd/prints/{printID}/mark_done").
		Queries("contest", "{contestID}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertTokenInRequest(t, r, token)
			vars := mux.Vars(r)
			doneids = append(doneids, vars["printID"])
			assert.Equal(t, contestid, vars["contestID"])
		})

	srv := httptest.NewServer(r)

	ctx := context.Background()
	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = token
	cfg.Toph.ContestID = contestid
	cfg.Debug.DontPrint = true
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
			pog:           pog.NewPogger(io.Discard, "", 0),
			delayNotFound: 125 * time.Millisecond,
		}.Loop(ctx)
	}()

	<-queue.emptyCh

	close(exitch)

	wg.Wait()

	select {
	case aborterr = <-abortch:
	default:
	}

	assert.Empty(t, queue.frames)

	return doneids, aborterr
}

func assertTokenInRequest(t *testing.T, r *http.Request, token string) {
	assert.Equal(t, fmt.Sprintf("Printd %s", token), r.Header.Get("Authorization"))
}

type Queue struct {
	frames  []Frame
	emptyCh chan struct{}
}

type Frame struct {
	print         *Print
	contestLocked bool
}

func NewQueue(frames []Frame) *Queue {
	q := &Queue{
		frames:  frames,
		emptyCh: make(chan struct{}),
	}
	if len(q.frames) == 0 {
		close(q.emptyCh)
	}
	return q
}

func (q *Queue) Next() *Frame {
	if len(q.frames) == 0 {
		return nil
	}
	fr := q.frames[0]
	q.frames = q.frames[1:]
	if len(q.frames) == 0 {
		close(q.emptyCh)
	}
	return &fr
}
