package main

import (
	"encoding/xml"
	"net/http"
	"os"

	"github.com/upgear/gokit/log"
	"github.com/upgear/gokit/web"
)

func main() {
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Info("listening for http traffic", log.KV{"port": port})
	log.Fatal(http.ListenAndServe(":"+port,
		web.Logger(http.HandlerFunc(handle)),
	))
}

func handle(w http.ResponseWriter, r *http.Request) {
	rt := web.NewResponseType(w, r)

	switch r.URL.Path {
	case "/400":
		// This 4XX message will be logged and sent to the client
		rt.Error("ut oh", http.StatusBadRequest)
		return
	case "/500":
		// This 5XX message will be logged but a generic message will be sent
		// to the client
		rt.Error("ohhhh noooo", http.StatusInternalServerError)
		return
	default:
		rt.WriteHeader(http.StatusOK)
		rt.Encode(struct {
			XMLName xml.Name `json:"-" xml:"response"`
			FooBar  string   `json:"foo" xml:"bar"`
		}{FooBar: "it is all 200 ok"})
	}
}
