package web_test

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/upgear/go-kit/web"
)

func Example_DoRetry() {
	req, err := http.NewRequest("GET", "https://some-api", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Try to 3 times to get a 2XX back
	resp, err := web.DefaultClient().Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func Example_DoUnmarshal() {
	req, err := http.NewRequest("GET", "https://some-api", nil)
	if err != nil {
		log.Fatal(err)
	}

	var response struct {
		ABC int `json:"abc" xml:"abc"`
	}

	if _, err := web.DefaultClient().DoUnmarshal(req, &response); err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.ABC)
}

func Example_handler() {
	web.GlobalContentTypePolicy = web.ContentTypePolicyJSONOrXML

	http.ListenAndServe(":8080", web.Logware(web.Contentware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				ABC int `json:"abc" xml:"abc"`
			}
			if err := web.RequestDecoder(r).Decode(&request); err != nil {
				web.Error(w, err, http.StatusBadRequest)
				return
			}

			switch request.ABC {
			case 400:
				// This 4XX message will be logged and sent to the client
				web.Error(w, errors.New("ut oh"), http.StatusBadRequest)
			case 500:
				// This 5XX message will be logged but a generic message will be sent
				// to the client
				web.Error(w, errors.New("ohhhh noooo"), http.StatusInternalServerError)
			default:
				type response struct {
					FooBar int `json:"foo" xml:"bar"`
				}
				// Serialize the struct to JSON or XML based on the `Accept` header
				web.Respond(w, response{FooBar: request.ABC}, http.StatusOK)
			}
		}),
	)))
}
