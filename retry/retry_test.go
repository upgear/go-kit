package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/upgear/go-kit/retry"
)

func TestRun(t *testing.T) {
	var i int

	(&retry.Policy{Attempts: 3, Sleep: time.Nanosecond, Factor: 2}).Run(
		func() error {
			i++
			return errors.New("ut oh")
		})

	if exp := 3; i != exp {
		t.Fatalf("expected exactly %v tries, got: %v", exp, i)
	}
}

func TestRunNil(t *testing.T) {
	var i int

	(&retry.Policy{Attempts: 3, Sleep: time.Nanosecond, Factor: 2}).Run(
		func() error {
			i++
			return nil
		})

	if exp := 1; i != exp {
		t.Fatalf("expected exactly %v tries, got: %v", exp, i)
	}
}

func TestStop(t *testing.T) {
	var i int

	(&retry.Policy{Attempts: 3, Sleep: time.Nanosecond, Factor: 2}).Run(
		func() error {
			i++
			return retry.Stop(errors.New("ut oh"))
		})

	if exp := 1; i != exp {
		t.Fatalf("expected exactly %v tries, got: %v", exp, i)
	}
}

func TestWithContext(t *testing.T) {
	fn := func() error {
		// Hang
		<-make(chan struct{})
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := retry.Double(3).Run(retry.WithContext(ctx, fn))

	if err != context.Canceled {
		t.Fatalf("expected err %q, got: %q", context.Canceled, err)
	}
}
