package web

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/upgear/go-kit/retry"
)

func DoRetry(attempts int, c *http.Client, req *http.Request) (*http.Response, error) {
	var resp *http.Response

	p := retry.Policy{Attempts: attempts, Sleep: time.Second, Factor: 2}

	nonReturnErr := errors.New("")

	err := retry.Run(p, func() error {
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
	if x, err := strconv.Atoi(h); err != nil {
		p.Sleep = time.Duration(time.Duration(x) * time.Second)
		return
	}
}
