// Package retry implements basic helpers for implementing retry policies.
//
// If you are looking to implement a retry policy in an http client, take a
// look at the web package.
package retry

import "time"

// Policy specifies how to execute Run(...).
type Policy struct {
	// Attempts to retry
	Attempts int
	// Sleep is the initial duration to wait before retrying
	Sleep time.Duration
	// Factor is the backoff rate (2 = double sleep time before next attempt)
	Factor int
}

// Double is a convenience function for calling Run with
// Policy.Sleep = time.Second & Policy.Factor = 2.
func Double(attempts int, f func() error) error {
	return Run(&Policy{
		Attempts: attempts,
		Factor:   2,
		Sleep:    time.Second,
	}, f)
}

// Run executing a function until:
// - A nil error is returned
// - The max number of attempts has been reached
// - A Stop() wrapped error is returned
func Run(p *Policy, f func() error) error {
	if err := f(); err != nil {
		if _, ok := err.(stop); ok {
			return err
		}

		p.Attempts = p.Attempts - 1
		if p.Attempts > 0 {
			time.Sleep(p.Sleep)
			p.Sleep = time.Duration(p.Factor) * p.Sleep
			return Run(p, f)
		}
		return err
	}

	return nil
}

// Stop wraps an error returned by a retry func and stops subsequent retries.
func Stop(err error) error {
	return stop{err}
}

type stop struct {
	error
}
