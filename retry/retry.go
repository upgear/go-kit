// Package retry implements basic helpers for implementing retry policies.
//
// If you are looking to implement a retry policy in an http client, take a
// look at the web package.
package retry

import (
	"context"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Stop wraps an error returned by a retry func and stops subsequent retries.
func Stop(err error) error {
	return stop{err}
}

type stop struct {
	error
}

// Policy specifies how to execute Run(...).
type Policy struct {
	// Attempts to retry
	Attempts int
	// Sleep is the initial duration to wait before retrying
	Sleep time.Duration
	// Factor is the backoff rate (2 = double sleep time before next attempt)
	Factor int
}

// Double is a convenience Policy which has a initial Sleep of 1 second and
// doubles every subsequent attempt.
func Double(attempts int) *Policy {
	return &Policy{
		Attempts: attempts,
		Factor:   2,
		Sleep:    time.Second,
	}
}

// Run executes a function until:
// 1. A nil error is returned,
// 2. The max number of attempts has been reached,
// 3. A Stop(...) wrapped error is returned
func (p *Policy) Run(f func() error) error {
	if err := f(); err != nil {
		if s, ok := err.(stop); ok {
			// Return the original error for later checking
			return s.error
		}

		p.Attempts = p.Attempts - 1
		if p.Attempts > 0 {
			p.sleep()
			return p.Run(f)
		}
		return err
	}

	return nil
}

// WithContext wraps a run function with a function that will return early
// if the context is done. In such a case Stop(ctx.Err()) is returned rather
// than the function's return value.
func WithContext(ctx context.Context, f func() error) func() error {
	return func() error {
		c := make(chan error, 1)

		go func() { c <- f() }()

		select {
		case <-ctx.Done():
			return Stop(ctx.Err())
		case err := <-c:
			return err
		}
	}
}

func (p *Policy) sleep() {
	// Add some randomness to prevent creating a Thundering Herd
	jitter := time.Duration(rand.Int63n(int64(p.Sleep)))
	p.Sleep = p.Sleep + jitter/time.Duration(p.Factor)

	time.Sleep(p.Sleep)
	p.Sleep = time.Duration(p.Factor) * p.Sleep
}
