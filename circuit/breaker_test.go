package circuit_test

import (
	"errors"
	"testing"
	"time"

	"github.com/upgear/go-kit/circuit"
)

func TestRunNoErrs(t *testing.T) {
	b := circuit.NewBreaker(3, time.Second)

	f := func() error {
		return nil
	}

	if err := b.Run(f); err != nil {
		t.Fatal("expected nil error")
	}
}

func TestRunErrs(t *testing.T) {
	// A breaker that waits a year before switching to half-open
	b := circuit.NewBreaker(3, time.Hour*24*365)

	myErr := errors.New("whoops")

	var n int
	f := func() error {
		n++
		return myErr
	}

	for i := 0; i <= 3; i++ {
		err := b.Run(f)
		if i == 3 {
			if err != circuit.ErrBreakerOpen {
				t.Fatalf("expected %s, got: %s", circuit.ErrBreakerOpen, err)
			}
		} else {
			if err != myErr {
				t.Fatalf("expected %s, got: %s", myErr, err)
			}
		}
	}

	if exp := 3; n != exp {
		t.Fatalf("expected %v runs, got %v", exp, n)
	}
}

func TestRunErrsTimeout(t *testing.T) {
	b := circuit.NewBreaker(3, time.Millisecond)

	myErr := errors.New("whoops")

	var n int
	f := func() error {
		n++
		return myErr
	}

	for i := 0; i <= 3; i++ {
		err := b.Run(f)
		if err != myErr {
			t.Fatalf("expected %s, got: %s", myErr, err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	if exp := 4; n != exp {
		t.Fatalf("expected %v runs, got %v", exp, n)
	}
}
