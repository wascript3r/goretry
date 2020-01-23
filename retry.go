package goretry

import (
	"context"
	"time"
)

type RetryFunc func() error

func Do(ctx context.Context, f RetryFunc, intv time.Duration, now bool) {
	go retry(ctx, f, intv, now)
}

func DoBlock(ctx context.Context, f RetryFunc, intv time.Duration, now bool) {
	retry(ctx, f, intv, now)
}

func retry(ctx context.Context, f RetryFunc, intv time.Duration, now bool) {
	if ctx.Err() == context.Canceled {
		return
	}

	if now {
		if err := f(); err == nil {
			return
		}
	}
	t := time.NewTimer(intv)

	for {
		select {
		case <-ctx.Done():
			t.Stop()
			return

		case <-t.C:
			if err := f(); err == nil {
				return
			}
			t.Reset(intv)
		}
	}
}
