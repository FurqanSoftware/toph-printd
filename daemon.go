package main

import (
	"context"
	"errors"
	"time"

	"github.com/FurqanSoftware/pog"
	"github.com/avast/retry-go"
)

type Daemon struct {
	cfg           Config
	params        Parameters
	exitCh        chan struct{}
	abortCh       chan error
	pog           *pog.Pogger
	delayNotFound time.Duration
}

func (d Daemon) Loop(ctx context.Context) {
L:
	for {
		stop, delay := d.iter(ctx)
		if stop {
			break
		}

		select {
		case <-d.exitCh:
			break L
		case <-time.After(delay):
		}
	}
}

func (d Daemon) iter(ctx context.Context) (stop bool, delay time.Duration) {
	pr, err := getNextPrint(ctx, d.cfg)
	var terr tophError
	if errors.As(err, &terr) {
		d.pog.SetStatus(statusOffline)
		d.pog.Error(err)
		if !errors.As(err, &retryableError{}) {
			d.abortCh <- err
			return true, 0
		}
		return false, d.cfg.Printd.DelayError
	}
	var perr noNextPrintError
	if errors.As(err, &perr) {
		if perr.contestLocked {
			d.pog.Info("Contest is locked")
			d.abortCh <- err
			return true, 0
		}
		d.pog.SetStatus(statusReady)
		return false, d.delayNotFound
	}
	catch(err)

	d.pog.SetStatus(statusPrinting)

	d.pog.Infof("Printing %s", pr.ID)
	d.pog.Infof("∟ Header: %s", pr.Header)
	if pr.PageLimit != -1 {
		d.pog.Infof("∟ Limit: %d", pr.PageLimit)
	}
	pdf, err := runPrintJob(ctx, d.cfg, pr)
	catch(err)
	err = retry.Do(func() error {
		return markPrintDone(ctx, d.cfg, pr, pdf)
	},
		retry.RetryIf(func(err error) bool { return errors.As(err, &retryableError{}) }),
		retry.Attempts(3),
		retry.Delay(500*time.Millisecond),
		retry.LastErrorOnly(true),
	)
	if errors.As(err, &terr) {
		d.pog.SetStatus(statusOffline)
		d.pog.Error(err)
		if !errors.As(err, &retryableError{}) {
			d.abortCh <- err
			return true, 0
		}
		return false, d.cfg.Printd.DelayError
	}
	catch(err)
	if pdf.PageSkipped > 0 {
		d.pog.Infof("∟ Pages: %d (Skipped: %d)", pdf.PageCount, pdf.PageSkipped)
	} else {
		d.pog.Infof("∟ Pages: %d", pdf.PageCount)
	}
	d.pog.Info("∟ Done")

	return false, d.cfg.Printd.DelayAfter
}
