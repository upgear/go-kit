package web_test

import (
	"encoding/xml"
	"errors"
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

func Example_response() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := web.NewResponse(w, r)

		switch r.URL.Path {
		case "/400":
			// This 4XX message will be logged and sent to the client
			res.SendErr(http.StatusBadRequest, errors.New("ut oh"))
		case "/500":
			// This 5XX message will be logged but a generic message will be sent
			// to the client
			res.SendErr(http.StatusInternalServerError, errors.New("ohhhh noooo"))
		default:
			type body struct {
				XMLName xml.Name `json:"-" xml:"response"`
				FooBar  string   `json:"foo" xml:"bar"`
			}
			// Serialize the struct to JSON or XML based on the `Accept` header
			res.Send(http.StatusOK, body{FooBar: "it is all 200 ok"})
		}
	}))
}
