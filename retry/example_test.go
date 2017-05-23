package retry_test

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/upgear/go-kit/retry"
)

func Example() {
	var resp *http.Response

	err := retry.Run(retry.Double(3), func() error {
		var err error
		resp, err = http.Get("https://golang.org")
		return err
	})

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
}

func Example_stop() {
	retry.Run(retry.Double(3), func() error {
		err := errors.New("no more retries after this error")

		return retry.Stop(err)
	})
}
