package web_test

import (
	"log"
	"net/http"

	"github.com/upgear/go-kit/web"
)

func Example_redo() {
	req, err := http.NewRequest("GET", "https://golang.org", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Try to 3 times to get a 2XX back
	resp, err := web.Redo(3, http.DefaultClient, req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}
