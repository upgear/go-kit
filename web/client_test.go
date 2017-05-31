package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/upgear/go-kit/web"
)

func TestDoRetry(t *testing.T) {
	var i int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch i {
		case 0: // 0 Seconds to get here
			w.WriteHeader(503)
		case 1: // ~1 Second to get here
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
		case 2: // ~1 More second to get here b/c of the previous `Retry-After` header
			w.WriteHeader(201)
		default:
			t.Fatal("this point should never have been reached")
		}
		i++
	}))

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("unable to make request: %s", err)
	}

	start := time.Now()

	const attempts = 3
	resp, err := web.DoRetry(http.DefaultClient, req, attempts)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if dur := time.Now().Sub(start); time.Second > dur || dur > 3*time.Second {
		t.Fatalf("expected DoRetry to take roughly 2 seconds, took: %s", dur)
	}

	if exp := 201; resp.StatusCode != exp {
		t.Fatalf("expected status %v, got: %v", exp, resp.StatusCode)
	}

	if i != attempts {
		t.Fatalf("expected %v attempts, got: %v", attempts, i)
	}
}

func TestDoUnmarshalJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"abc":123}`))
	}))

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	var response struct {
		ABC int `json:"abc" xml:"-"`
	}

	if _, err := web.DoUnmarshal(http.DefaultClient, req, &response); err != nil {
		t.Fatal(err)
	}

	if response.ABC != 123 {
		t.Fatal("incorrect unmarshalling")
	}
}

func TestDoUnmarshalXML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<response><abc>123</abc></response>`))
	}))

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	var response struct {
		ABC int `json:"-" xml:"abc"`
	}

	if _, err := web.DoUnmarshal(http.DefaultClient, req, &response); err != nil {
		t.Fatal(err)
	}

	if response.ABC != 123 {
		t.Fatal("incorrect unmarshalling")
	}
}

func TestDoUnmarshal5XX(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	var response struct{}
	resp, err := web.DoUnmarshal(http.DefaultClient, req, response)
	if err != web.Err5XX {
		t.Fatalf("expected Err5XX, got: %s", err)
	}
	if resp != nil {
		t.Fatal("expected nil response variable")
	}
}

func TestDoRetry5XX(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := web.DoRetry(http.DefaultClient, req, 3)
	if err != web.Err5XX {
		t.Fatalf("expected Err5XX, got: %s", err)
	}
	if resp != nil {
		t.Fatal("expected nil response variable")
	}
}
