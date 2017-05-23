package web_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/upgear/go-kit/web"
)

func TestDoRetry(t *testing.T) {
	var i int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch i {
		case 0:
			w.WriteHeader(503)
		case 1:
			w.WriteHeader(429)
		case 2:
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

	const attempts = 3
	resp, err := web.Redo(attempts, http.DefaultClient, req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if exp := 201; resp.StatusCode != exp {
		t.Fatalf("expected status %v, got: %v", exp, resp.StatusCode)
	}

	if i != attempts {
		t.Fatalf("expected %v attempts, got: %v", attempts, i)
	}
}
