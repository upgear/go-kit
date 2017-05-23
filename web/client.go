package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/upgear/go-kit/retry"
)

// Redo acts the same as http.Client.Do except it retries for any errors
// or status codes 420, 429, and 5XX.
func Redo(attempts int, c *http.Client, req *http.Request) (*http.Response, error) {
	var resp *http.Response

	p := retry.Policy{Attempts: attempts, Sleep: time.Second, Factor: 2}

	nonReturnErr := errors.New("")

	err := retry.Run(&p, func() error {
		var err error
		resp, err = c.Do(req)
		if err != nil {
			return err
		}

		if s := resp.StatusCode; s >= 500 || s == 420 || s == 429 {
			// Respect `Retry-After` headers
			alterPolicyFromRetryHeader(&p, resp.Header.Get("Retry-After"))

			// Retry again
			return nonReturnErr
		} else if s >= 400 {
			// Don't retry
			return retry.Stop(nonReturnErr)
		}

		return err
	})

	switch err {
	// Don't return errors http.Client.Do would not return
	case nonReturnErr, retry.Stop(nonReturnErr):
		return resp, nil
	default:
		return resp, err
	}
}

func alterPolicyFromRetryHeader(p *retry.Policy, h string) {
	// Seconds Variation: `Retry-After: 120`
	if x, err := strconv.ParseInt(h, 10, 64); err == nil {
		p.Sleep = time.Duration(time.Duration(x) * time.Second)
		return
	}
	// TODO: Implement Timestamp variation: `Retry-After: Fri, 31 Dec 1999 23:59:59 GMT`
}
