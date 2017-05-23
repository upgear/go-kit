package retry_test

import (
	"errors"
	"testing"
	"time"

	"github.com/upgear/go-kit/retry"
)

func TestRun(t *testing.T) {
	var i int

	retry.Run(&retry.Policy{Attempts: 3, Sleep: time.Nanosecond, Factor: 2},
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

	retry.Run(&retry.Policy{Attempts: 3, Sleep: time.Nanosecond, Factor: 2},
		func() error {
			i++
			return nil
		})

	if exp := 1; i != exp {
		t.Fatalf("expected exactly %v tries, got: %v", exp, i)
	}
}

func TestDouble(t *testing.T) {
	var i int

	retry.Double(3, func() error {
		i++
		return errors.New("ut oh")
	})

	if exp := 3; i != exp {
		t.Fatalf("expected exactly %v tries, got: %v", exp, i)
	}
}

func TestStop(t *testing.T) {
	var i int

	retry.Run(&retry.Policy{Attempts: 3, Sleep: time.Nanosecond, Factor: 2},
		func() error {
			i++
			return retry.Stop(errors.New("ut oh"))
		})

	if exp := 1; i != exp {
		t.Fatalf("expected exactly %v tries, got: %v", exp, i)
	}
}
