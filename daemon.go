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
	delay := 0 * time.Second
L:
	for {
		pr, err := getNextPrint(ctx, d.cfg)
		var terr tophError
		if errors.As(err, &terr) {
			d.pog.SetStatus(statusOffline)
			d.pog.Error(err)
			if !errors.As(err, &retryableError{}) {
				d.abortCh <- err
				break L
			}
			delay = d.cfg.Printd.DelayError
			goto retry
		}
		catch(err)

		if pr.ID == "" {
			d.pog.SetStatus(statusReady)
			delay = d.delayNotFound
			goto retry
		}

		d.pog.SetStatus(statusPrinting)

		d.pog.Infof("Printing %s", pr.ID)
		err = runPrintJob(ctx, d.cfg, pr)
		catch(err)
		err = retry.Do(func() error {
			return markPrintDone(ctx, d.cfg, pr)
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
				break L
			}
			delay = d.cfg.Printd.DelayError
			goto retry
		}
		catch(err)
		d.pog.Info("∟ Done")

		delay = d.cfg.Printd.DelayAfter

	retry:
		select {
		case <-d.exitCh:
			break L
		case <-time.After(delay):
		}
	}
}
