package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/upgear/go-kit/retry"
)

var Err4XX = errors.New("4XX: client error")
var Err5XX = errors.New("5XX: server error")

// DoRetry acts the same as http.Client.Do except it retries for any errors
// or status codes 420, 429, and 5XX.
// Any 4XX or 5XX statuses will return an error with a non-nil response value.
func DoRetry(c *http.Client, r *http.Request, attempts int) (*http.Response, error) {
	var resp *http.Response

	p := retry.Double(attempts)

	err := retry.Run(p, func() error {
		var err error
		resp, err = c.Do(r)
		if err != nil {
			return err
		}

		if s := resp.StatusCode; s >= 500 || s == 420 || s == 429 {
			// Respect `Retry-After` headers
			alterPolicyFromRetryHeader(p, resp.Header.Get("Retry-After"))

			// Retry again
			if s >= 500 {
				return Err5XX
			} else {
				return Err4XX
			}
		} else if s >= 400 {
			// Don't retry
			return retry.Stop(Err4XX)
		}

		return err
	})

	return resp, err
}

func alterPolicyFromRetryHeader(p *retry.Policy, h string) {
	// Seconds Variation: `Retry-After: 120`
	if x, err := strconv.ParseInt(h, 10, 64); err == nil {
		p.Sleep = time.Duration(time.Duration(x) * time.Second)
		return
	}
	// TODO: Implement Timestamp variation: `Retry-After: Fri, 31 Dec 1999 23:59:59 GMT`
}

// DoUnmarshal makes an http request and attempts to unmarshal the response.
// Any 4XX or 5XX statuses will return an error with a non-nil response value.
//
// It will attempt to unmarshal any 2XX responses and close the
// http.Response.Body.
func DoUnmarshal(c *http.Client, r *http.Request, x interface{}) (*http.Response, error) {
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}

	if s := resp.StatusCode; s >= 500 {
		return resp, Err5XX
	} else if s >= 400 {
		return resp, Err4XX
	}

	defer resp.Body.Close()
	if err := ResponseDecoder(resp).Decode(x); err != nil {
		return nil, errors.Wrap(err, "unable to decode response")
	}

	return resp, nil
}
