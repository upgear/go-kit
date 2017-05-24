package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/upgear/go-kit/log"
)

type ContentTypePolicy uint32
type ContentType string

const (
	ContentTypePolicyJSONOrXML = ContentTypePolicy(iota)
	ContentTypePolicyJSONOnly
	ContentTypePolicyXMLOnly
)

var (
	// GlobalContentTypePolicy is a global variable that should not be changed
	// while handlers are active. It defaults to ContentTypePolicyJSONOrXML
	// but is set at init() time using the `HTTP_CONTENT_TYPE` env variable.
	GlobalContentTypePolicy ContentTypePolicy

	ContentTypeXML  = ContentType("application/xml")
	ContentTypeJSON = ContentType("application/json")
)

func init() {
	switch strings.ToLower(os.Getenv("HTTP_CONTENT_TYPE")) {
	case "json":
		GlobalContentTypePolicy = ContentTypePolicyJSONOnly
	case "xml":
		GlobalContentTypePolicy = ContentTypePolicyXMLOnly
	default:
		GlobalContentTypePolicy = ContentTypePolicyJSONOrXML
	}
}

type Decoder interface {
	Decode(interface{}) error
}

// RequestDecoder wraps a http.Request.Body with either a JSON or XML encoder
// based on the request `Content-Type` header.
func RequestDecoder(r *http.Request) Decoder {
	return newDecoder(r.Header.Get("Content-Type"), r.Body)
}

// ResponseDecoder wraps a http.Response.Body with the either a JSON or XML encoder
// based on the response `Content-Type` header.
func ResponseDecoder(r *http.Response) Decoder {
	return newDecoder(r.Header.Get("Content-Type"), r.Body)
}

func newDecoder(ct string, rdr io.Reader) Decoder {
	switch jsonOrXML(ct) {
	case ContentTypeXML:
		return xml.NewDecoder(rdr)
	default:
		return json.NewDecoder(rdr)
	}
}

type Encoder interface {
	Encode(interface{}) error
}

// ResponseEncoder wraps a http.ResponseWriter with the either a JSON or XML encoder
// based on the previously set response `Content-Type` header.
func ResponseEncoder(w http.ResponseWriter) Encoder {
	h := w.Header().Get("Content-Type")

	switch jsonOrXML(h) {
	case ContentTypeXML:
		return json.NewEncoder(w)
	default:
		return xml.NewEncoder(w)
	}
}

func jsonOrXML(h string) ContentType {
	switch GlobalContentTypePolicy {
	case ContentTypePolicyJSONOnly:
		return ContentTypeJSON
	case ContentTypePolicyXMLOnly:
		return ContentTypeXML
	default:
		if strings.Contains(h, "xml") {
			return ContentTypeXML
		}
		return ContentTypeJSON
	}
}

// Respond encodes an JSON or XML response after writing a status code.
//
// This function logs.
func Respond(w http.ResponseWriter, x interface{}, status int) {
	enc := ResponseEncoder(w)

	w.WriteHeader(status)

	if err := enc.Encode(x); err != nil {
		log.Warn(errors.Wrap(err, "unable to marshal response to content type"))
	}
}

// Error writes a standardized error struct. For 5XX errors, the message sent
// across the wire will be generic while the real error message is logged.
// If err is nil, the standardized status text for the given status will be
// used as the message.
//
// This function logs.
func Error(w http.ResponseWriter, err error, status int) {
	if err == nil {
		err = errors.New(http.StatusText(status))
	}

	if status >= 500 {
		// Override outgoing message as to not display internal errors externally
		log.Error(err, log.M{"status": status})
		err = errors.New(http.StatusText(status))
		if err.Error() == "" {
			err = errors.New(http.StatusText(http.StatusInternalServerError))
		}
	} else {
		log.Warn(err, log.M{"status": status})
	}

	fmt.Println("err", err)
	Respond(
		w,
		struct {
			XMLName xml.Name `json:"-" xml:"error"`
			Error   string   `json:"error" xml:"message"`
		}{Error: err.Error()},
		status,
	)
}
