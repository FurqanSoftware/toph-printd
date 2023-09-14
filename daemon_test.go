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
	donePrintIDs := testDaemon(t, "ZnZ0eW81bWw1NjMzc2dlYjI5azVuNThleDVzYjZ1aG8=", "6502f105025832238e865526", NewQueue([]*Print{
		{
			ID:      "6502f46b17592a5a9e870928",
			Header:  "Test Print 1",
			Content: "Lorem ipsum dolor.",
			Status:  "Queued",
		},
		{
			ID:      "6502f9bf92c75c2f7874698e",
			Header:  "Test Print 2",
			Content: "The quick brown fox...",
			Status:  "Queued",
		},
	}))

	assert.FileExists(t, "6502f46b17592a5a9e870928.pdf")
	os.Remove("6502f46b17592a5a9e870928.pdf")
	assert.FileExists(t, "6502f9bf92c75c2f7874698e.pdf")
	os.Remove("6502f9bf92c75c2f7874698e.pdf")

	assert.Equal(t, []string{
		"6502f46b17592a5a9e870928",
		"6502f9bf92c75c2f7874698e",
	}, donePrintIDs)
}

func TestDaemonEmptyHeader(t *testing.T) {
	donePrintIDs := testDaemon(t, "dGlsa2c0em5lOG4waTl3MTl1d2prejVldmk5cGIycTU=", "6502fee52e058f9990cc5c6e", NewQueue([]*Print{
		{
			ID:      "6502fec79eed503a402e0b59",
			Header:  "",
			Content: "Lorem ipsum dolor.",
			Status:  "Queued",
		},
	}))

	assert.FileExists(t, "6502fec79eed503a402e0b59.pdf")
	os.Remove("6502fec79eed503a402e0b59.pdf")

	assert.Equal(t, []string{
		"6502fec79eed503a402e0b59",
	}, donePrintIDs)
}

func TestDaemonEmptyContent(t *testing.T) {
	donePrintIDs := testDaemon(t, "bWJocHBqMXA1aHRxbHpmMWo2bnNvYmhmcmVpYXlvdGQ=", "6502fee1c7bb2bf7288be9c7", NewQueue([]*Print{
		{
			ID:      "6502fecc74093fd44c060e11",
			Header:  "Keyboard Cat",
			Content: "",
			Status:  "Queued",
		},
	}))

	assert.FileExists(t, "6502fecc74093fd44c060e11.pdf")
	os.Remove("6502fecc74093fd44c060e11.pdf")

	assert.Equal(t, []string{
		"6502fecc74093fd44c060e11",
	}, donePrintIDs)
}

func TestDaemonBreak(t *testing.T) {
	donePrintIDs := testDaemon(t, "em95bHBmMTduM25wcDJyeTVucGd1bDI3MXB1d2V6ODM=", "6502fedddafbfc6ac0f20876", NewQueue([]*Print{
		{
			ID:      "6502fed2ee36d5244aade158",
			Header:  "Test Print 1",
			Content: "Lorem ipsum dolor.",
			Status:  "Queued",
		},
		nil,
		{
			ID:      "6502fed88ea6c9620b82bf5a",
			Header:  "Test Print 2",
			Content: "The quick brown fox...",
			Status:  "Queued",
		},
	}))

	assert.FileExists(t, "6502fed2ee36d5244aade158.pdf")
	os.Remove("6502fed2ee36d5244aade158.pdf")
	assert.FileExists(t, "6502fed88ea6c9620b82bf5a.pdf")
	os.Remove("6502fed88ea6c9620b82bf5a.pdf")

	assert.Equal(t, []string{
		"6502fed2ee36d5244aade158",
		"6502fed88ea6c9620b82bf5a",
	}, donePrintIDs)
}

func TestDaemonEmpty(t *testing.T) {
	donePrintIDs := testDaemon(t, "ZWV5aTFwc3hjNnc2c2NlZG13MHpreHUzaDc3cXhyMmg=", "6503070f91fb17000cc2e5b9", NewQueue([]*Print{}))

	assert.Equal(t, []string{}, donePrintIDs)
}

func testDaemon(t *testing.T, token, contestid string, queue *Queue) (donePrintIDs []string) {
	pog.InitDefault()

	r := mux.NewRouter()

	r.HandleFunc("/api/printd/contests/{contestID}/next_print", func(w http.ResponseWriter, r *http.Request) {
		assertTokenInRequest(t, r, token)
		pr := queue.Next()
		if pr == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		err := json.NewEncoder(w).Encode(pr)
		assert.NoError(t, err)
	})

	donePrintIDs = []string{}
	r.NewRoute().
		Path("/api/printd/prints/{printID}/mark_done").
		Queries("contest", "{contestID}").
		HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertTokenInRequest(t, r, token)
			vars := mux.Vars(r)
			donePrintIDs = append(donePrintIDs, vars["printID"])
			assert.Equal(t, contestid, vars["contestID"])
		})

	srv := httptest.NewServer(r)

	ctx := context.Background()
	cfg := Config{}
	cfg.initDefaults()
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.Token = token
	cfg.Toph.ContestID = contestid
	exitCh := make(chan struct{})
	abortCh := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		Daemon{
			cfg:           cfg,
			exitCh:        exitCh,
			abortCh:       abortCh,
			pog:           pog.NewPogger(io.Discard, "", 0),
			delayNotFound: 125 * time.Millisecond,
		}.Loop(ctx)
	}()

	select {
	case <-queue.EmptyCh:
	case <-abortCh:
	}
	close(exitCh)

	wg.Wait()

	assert.Empty(t, queue.Prints)

	return donePrintIDs
}

func assertTokenInRequest(t *testing.T, r *http.Request, token string) {
	assert.Equal(t, fmt.Sprintf("Printd %s", token), r.Header.Get("Authorization"))
}

type Queue struct {
	Prints  []*Print
	EmptyCh chan struct{}
}

func NewQueue(prints []*Print) *Queue {
	q := &Queue{
		Prints:  prints,
		EmptyCh: make(chan struct{}),
	}
	if len(q.Prints) == 0 {
		close(q.EmptyCh)
	}
	return q
}

func (q *Queue) Next() *Print {
	if len(q.Prints) == 0 {
		return nil
	}
	pr := q.Prints[0]
	q.Prints = q.Prints[1:]
	if len(q.Prints) == 0 {
		close(q.EmptyCh)
	}
	return pr
}
