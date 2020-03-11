package goretry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInterrupted      = errors.New("retry function was interrupted by context")
	ErrAttemptsExceeded = errors.New("attempts limit is exceeded")
	ErrStopped          = errors.New("retry function was stopped")
)

type StopError struct {
	Err error
}

func (s *StopError) Error() string {
	if s.Err == nil {
		return ErrStopped.Error()
	}
	return fmt.Sprintf("%v: %v", ErrStopped, s.Err)
}

func UnwrapStopErr(err error) (error, bool) {
	if e, ok := err.(*StopError); ok {
		return e.Err, true
	}
	return err, false
}

type RetryFunc func() error

func Do(ctx context.Context, f RetryFunc, intv time.Duration, attempts int, now bool) {
	go retry(ctx, f, intv, attempts, now)
}

func DoBlock(ctx context.Context, f RetryFunc, intv time.Duration, attempts int, now bool) error {
	return retry(ctx, f, intv, attempts, now)
}

func StopErr(err error) error {
	return &StopError{err}
}

func ctxErr(ctx context.Context) error {
	if ctx.Err() != nil {
		return ErrInterrupted
	}
	return nil
}

func retry(ctx context.Context, f RetryFunc, intv time.Duration, attempts int, now bool) error {
	if e := ctxErr(ctx); e != nil {
		return e
	}

	callf := func() (error, bool) {
		err := f()
		if err == nil {
			return ctxErr(ctx), true
		}
		if _, ok := err.(*StopError); ok {
			return err, true
		}
		if e := ctxErr(ctx); e != nil {
			return e, true
		}
		return err, false
	}

	if now {
		err, exit := callf()
		if exit {
			return err
		}
	}
	t := time.NewTimer(intv)

	for i := attempts; i > 0; i-- {
		select {
		case <-ctx.Done():
			t.Stop()
			return ErrInterrupted

		case <-t.C:
			err, exit := callf()
			if exit {
				return err
			}
			t.Reset(intv)
		}
	}

	return ErrAttemptsExceeded
}
