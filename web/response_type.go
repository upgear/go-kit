package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/upgear/gokit/log"
)

type EncodeFunc func(interface{}) error

type ResponseType struct {
	ContentType string
	Encode      EncodeFunc
	w           http.ResponseWriter
}

func FromAccept(w http.ResponseWriter, accept string) ResponseType {
	if strings.Contains(accept, "xml") {
		return ResponseType{"application/xml", xml.NewEncoder(w).Encode, w}
	}
	return ResponseType{"application/json", json.NewEncoder(w).Encode, w}
}

func FromRequest(w http.ResponseWriter, r *http.Request) ResponseType {
	return FromAccept(w, r.Header.Get("Accept"))
}

func (rt ResponseType) WriteHeader(status int) {
	rt.w.Header().Set("Content-Type", rt.ContentType)
	rt.w.WriteHeader(status)
}

func (rt ResponseType) Error(err interface{}, status int) {
	rt.WriteHeader(status)

	if status >= 500 {
		// Override outgoing message as to not display internal errors externally
		log.Error(err, log.KV{"status": status})
		err = http.StatusText(status)
		if err == "" {
			err = http.StatusText(http.StatusInternalServerError)
		}
	} else {
		log.Warn(err, log.KV{"status": status})
	}

	if err := rt.Encode(struct {
		XMLName xml.Name `json:"-" xml:"error"`
		Error   string   `json:"error" xml:"message"`
	}{Error: fmt.Sprintf("%s", err)}); err != nil {
		log.Error("unable to write json to client: " + err.Error())
	}
}
