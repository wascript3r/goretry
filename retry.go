package goretry

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInterrupted = NewError(
		ExitState,
		errors.New("retry function was interrupted by context"),
	)
	ErrAttemptsExceeded = NewError(
		StopState,
		errors.New("attempts limit is exceeded"),
	)
)

type RetryFunc func() *Error

func Do(ctx context.Context, f RetryFunc, intv time.Duration, attempts int, now bool) {
	go retry(ctx, f, intv, attempts, now)
}

func DoBlock(ctx context.Context, f RetryFunc, intv time.Duration, attempts int, now bool) *Error {
	return retry(ctx, f, intv, attempts, now)
}

func ctxErr(ctx context.Context) *Error {
	if ctx.Err() != nil {
		return ErrInterrupted
	}
	return nil
}

func retry(ctx context.Context, f RetryFunc, intv time.Duration, attempts int, now bool) *Error {
	if e := ctxErr(ctx); e != nil {
		return e
	}

	callf := func() *Error {
		err := f()
		if err == nil {
			return ctxErr(ctx)
		}

		if !IsValidState(err.State) {
			err.State = InvalidState
			return err
		}

		if err.State != ContinueState {
			return err
		}
		if e := ctxErr(ctx); e != nil {
			return e
		}
		return err
	}

	if now {
		err := callf()
		if err == nil {
			return nil
		}
		if err.State != ContinueState {
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
			err := callf()
			if err == nil {
				return nil
			}
			if err.State != ContinueState {
				return err
			}
			t.Reset(intv)
		}
	}

	return ErrAttemptsExceeded
}
