package web

import (
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/upgear/go-kit/circuit"
	"github.com/upgear/go-kit/retry"
)

var Err4XX = errors.New("4XX client error")
var Err5XX = errors.New("5XX server error")

// DefaultClient is a function rather than a var (as in the http pkg) because
// it includes a circuit breaker so it should not be used as a client for
// multiple services.
func DefaultClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		RetryPolicy:    retry.Double(3),
		CircuitBreaker: circuit.NewBreaker(100, time.Second),
	}
}

// Client wraps http.Client while adding retry and circuit breaker
// functionality. Because it has an embedded circuit breaker, a single
// Client (with a non-nil CircuitBreaker) should not be used to connect to
// multiple services. A failure in one service would trip the breaker for other
// services.
type Client struct {
	HTTPClient *http.Client
	// RetryPolicy can be nil and a zero'd retry policy (aka 1 try will be used)
	RetryPolicy *retry.Policy
	// CircuitBreaker can be nil and it will be ignored.
	CircuitBreaker *circuit.Breaker
}

// Do acts the same as http.Client.Do except:
//
// - It retries for any errors or status codes 420, 429, and 5XX.
// - Circuit breaker functionality can be configured.
// - 4XX or 5XX statuses will return an error with a nil response value.
//
func (c *Client) Do(r *http.Request) (*http.Response, error) {
	p := c.RetryPolicy
	if p == nil {
		p = &retry.Policy{}
	}
	return c.do(r, *p, c.CircuitBreaker)
}

func (c *Client) do(r *http.Request, p retry.Policy, b *circuit.Breaker) (*http.Response, error) {
	var resp *http.Response

	// Define a function which maps http status codes to errors
	doHTTP := func() error {
		var err error
		resp, err = c.HTTPClient.Do(r)
		if err != nil {
			return err
		}

		s := resp.StatusCode
		switch {
		case s == 420 || s == 429:
			alterPolicyFromRetryHeader(&p, resp.Header.Get("Retry-After"))
			return wrapErrStatus(Err4XX, s)
		case s >= 500:
			alterPolicyFromRetryHeader(&p, resp.Header.Get("Retry-After"))
			return wrapErrStatus(Err5XX, s)
		case s >= 400:
			return retry.Stop(wrapErrStatus(Err4XX, s))
		default: // Success
			return nil
		}
	}

	fn := doHTTP

	// Wrap the function in a circuit breaker if one is defined
	if b != nil {
		fn = func() error {
			err := doHTTP()
			// Don't trip on client errors
			if errors.Cause(err) == Err4XX {
				err = circuit.Ignore(err)
			}
			return err
		}
	}

	err := p.Run(retry.WithContext(r.Context(), fn))

	if errors.Cause(err) == Err4XX || errors.Cause(err) == Err5XX {
		defer resp.Body.Close()
		return nil, wrapErrBody(err, resp.Body)
	}
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DoUnmarshal makes an http request and attempts to unmarshal the response.
// Any 4XX or 5XX statuses will return an error with a nil response value.
//
// It will attempt to unmarshal any 2XX responses.
//
// This function closes the http.Response.Body.
func (c *Client) DoUnmarshal(r *http.Request, x interface{}) (*http.Response, error) {
	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if s := resp.StatusCode; s >= 500 {
		return nil, wrapErrBody(wrapErrStatus(Err5XX, s), resp.Body)
	} else if s >= 400 {
		return nil, wrapErrBody(wrapErrStatus(Err4XX, s), resp.Body)
	}

	if err := ResponseDecoder(resp).Decode(x); err != nil {
		return nil, errors.Wrap(err, "unable to decode response")
	}

	return resp, nil
}

// alterPolicyFromRetryHeader adjusts a retry policy's sleep duration
// based on headers sent back from a server.
func alterPolicyFromRetryHeader(p *retry.Policy, h string) {
	// Seconds Variation: `Retry-After: 120`
	if x, err := strconv.ParseInt(h, 10, 64); err == nil {
		p.Sleep = time.Duration(time.Duration(x) * time.Second)
		return
	}
	// TODO: Implement Timestamp variation: `Retry-After: Fri, 31 Dec 1999 23:59:59 GMT`
}

// wrapErrBody wraps an HTTP error with the body to provide extra context
func wrapErrBody(err error, body io.ReadCloser) error {
	btys, _ := ioutil.ReadAll(body)
	return errors.Wrapf(err, "response body: %s", string(btys))
}

// wrapErrStatus wraps an HTTP error with context about the status code
func wrapErrStatus(err error, status int) error {
	return errors.Wrapf(err, "status code %v", status)
}
