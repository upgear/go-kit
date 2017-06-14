package web

import (
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/upgear/go-kit/retry"
)

var Err4XX = errors.New("4XX client error")
var Err5XX = errors.New("5XX server error")

var DefaultClient = Client{
	HTTPClient: &http.Client{
		Timeout: 10 * time.Second,
	},
	RetryPolicy: retry.Double(3),
}

func Do(r *http.Request) (*http.Response, error) {
	return DefaultClient.Do(r)
}

func DoUnmarshal(r *http.Request, x interface{}) (*http.Response, error) {
	return DefaultClient.DoUnmarshal(r, x)
}

type Client struct {
	HTTPClient  *http.Client
	RetryPolicy *retry.Policy
}

// Do acts the same as http.Client.Do except it retries for any errors
// or status codes 420, 429, and 5XX.
// Any 4XX or 5XX statuses will return an error with a nil response value.
func (c *Client) Do(r *http.Request) (*http.Response, error) {
	return c.doRetry(r, *c.RetryPolicy)
}

func (c *Client) doRetry(r *http.Request, p retry.Policy) (*http.Response, error) {
	var resp *http.Response

	err := p.Run(func() error {
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
	})

	if errors.Cause(err) == Err4XX || errors.Cause(err) == Err5XX {
		defer resp.Body.Close()
		return nil, wrapErrBody(err, resp.Body)
	}

	return resp, err
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
