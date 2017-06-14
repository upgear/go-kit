package circuit

import (
	"errors"
	"sync/atomic"
	"time"
)

var ErrBreakerOpen = errors.New("breaker open")

func NewBreaker(threshold int64, timeout time.Duration) *Breaker {
	return &Breaker{
		Threshold: threshold,
		Timeout:   timeout,
	}
}

type Breaker struct {
	// Threshold is the number of times consecutive failures may occur
	// before the breaker gets flipped.
	// NOTE: This variable is not safe to change while concurrently calling Run(...).
	Threshold int64
	// Timeout if the duration to wait before retrying after a flip or after
	// the last attempt in a half-open state.
	// NOTE: This variable is not safe to change while concurrently calling Run(...).
	Timeout time.Duration

	// failures is protected with atomic
	failures int64
	// timestamp (unix nanoseconds) is protected with atomic
	timestamp int64
}

func (b *Breaker) Run(f func() error) error {
	if !b.allowed() {
		return ErrBreakerOpen
	}

	if err := f(); err != nil {
		if e, ok := err.(ignore); ok {
			return e
		}

		b.fail()
		return err
	}

	b.succeed()
	return nil
}

func Ignore(err error) error {
	return ignore{err}
}

type ignore struct {
	error
}

func (b *Breaker) allowed() bool {
	fails := atomic.LoadInt64(&b.failures)

	if fails >= b.Threshold {
		since := time.Now().Sub(time.Unix(0, b.timestamp))
		if since > b.Timeout {
			// Allow one through and start the timer again
			b.setTimer()
			return true
		}
		return false
	}

	return true
}

func (b *Breaker) fail() {
	atomic.AddInt64(&b.failures, 1)
	b.setTimer()
}

func (b *Breaker) setTimer() {
	atomic.StoreInt64(&b.timestamp, time.Now().UnixNano())
}

func (b *Breaker) succeed() {
	atomic.StoreInt64(&b.failures, 0)
}
