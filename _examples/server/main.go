package main

import (
	"encoding/xml"
	"errors"
	"net/http"
	"os"

	"github.com/upgear/go-kit/log"
	"github.com/upgear/go-kit/web"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("listening for http traffic", log.M{"port": port})
	log.Fatal(http.ListenAndServe(":"+port,
		web.Logger(http.HandlerFunc(handle)),
	))
}

func handle(w http.ResponseWriter, r *http.Request) {
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
		res.Send(http.StatusOK, struct {
			XMLName xml.Name `json:"-" xml:"response"`
			FooBar  string   `json:"foo" xml:"bar"`
		}{FooBar: "it is all 200 ok"})
	}
}
