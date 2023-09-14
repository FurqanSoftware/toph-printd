package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/FurqanSoftware/pog"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPrintLoop(t *testing.T) {
	donePrintIDs := testPrintLoop(t, "6502f105025832238e865526", NewQueue([]*Print{
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
	assert.FileExists(t, "6502f9bf92c75c2f7874698e.pdf")

	assert.Equal(t, []string{
		"6502f46b17592a5a9e870928",
		"6502f9bf92c75c2f7874698e",
	}, donePrintIDs)
}

func TestPrintLoopEmptyHeader(t *testing.T) {
	donePrintIDs := testPrintLoop(t, "6502fee52e058f9990cc5c6e", NewQueue([]*Print{
		{
			ID:      "6502fec79eed503a402e0b59",
			Header:  "",
			Content: "Lorem ipsum dolor.",
			Status:  "Queued",
		},
	}))

	assert.FileExists(t, "6502fec79eed503a402e0b59.pdf")

	assert.Equal(t, []string{
		"6502fec79eed503a402e0b59",
	}, donePrintIDs)
}

func TestPrintLoopEmptyContent(t *testing.T) {
	donePrintIDs := testPrintLoop(t, "6502fee1c7bb2bf7288be9c7", NewQueue([]*Print{
		{
			ID:      "6502fecc74093fd44c060e11",
			Header:  "Keyboard Cat",
			Content: "",
			Status:  "Queued",
		},
	}))

	assert.FileExists(t, "6502fecc74093fd44c060e11.pdf")

	assert.Equal(t, []string{
		"6502fecc74093fd44c060e11",
	}, donePrintIDs)
}

func TestPrintLoopBreak(t *testing.T) {
	donePrintIDs := testPrintLoop(t, "6502fedddafbfc6ac0f20876", NewQueue([]*Print{
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
	assert.FileExists(t, "6502fed88ea6c9620b82bf5a.pdf")

	assert.Equal(t, []string{
		"6502fed2ee36d5244aade158",
		"6502fed88ea6c9620b82bf5a",
	}, donePrintIDs)
}

func testPrintLoop(t *testing.T, contestid string, queue *Queue) (donePrintIDs []string) {
	pog.InitDefault()

	r := mux.NewRouter()

	r.HandleFunc("/api/printd/contests/{contestID}/next_print", func(w http.ResponseWriter, r *http.Request) {
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
			vars := mux.Vars(r)
			donePrintIDs = append(donePrintIDs, vars["printID"])
			assert.Equal(t, contestid, vars["contestID"])
		})

	srv := httptest.NewServer(r)

	ctx := context.Background()
	cfg := Config{}
	cfg.initDefaults()
	cfg.Printd.KeepPDF = true
	cfg.Printer.PageSize = PageA4
	cfg.Toph.BaseURL = srv.URL
	cfg.Toph.ContestID = contestid
	exitCh := make(chan struct{})
	abortCh := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		printLoop(ctx, cfg, exitCh, abortCh)
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

type Queue struct {
	Prints  []*Print
	EmptyCh chan struct{}
}

func NewQueue(prints []*Print) *Queue {
	return &Queue{
		Prints:  prints,
		EmptyCh: make(chan struct{}),
	}
}

func (q *Queue) Next() *Print {
	pr := q.Prints[0]
	q.Prints = q.Prints[1:]
	if len(q.Prints) == 0 {
		close(q.EmptyCh)
	}
	return pr
}
