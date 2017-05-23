package web

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/upgear/go-kit/log"
)

type ContentTypePolicy uint32

const (
	ContentTypePolicyJSONOrXML = iota
	ContentTypePolicyJSONOnly
	ContentTypePolicyXMLOnly
)

var (
	// ResponseTypePolicy is a global variable that should not be changed
	// while handlers are active. It defaults to ContentTypePolicyJSONOrXML
	// but is set at init() time using the `HTTP_RESPONSE_TYPE` env variable.
	ResponseTypePolicy ContentTypePolicy

	ContentTypeXML  = "application/xml"
	ContentTypeJSON = "application/json"
)

func init() {
	switch strings.ToLower(os.Getenv("HTTP_RESPONSE_TYPE")) {
	case "json":
		ResponseTypePolicy = ContentTypePolicyJSONOnly
	case "xml":
		ResponseTypePolicy = ContentTypePolicyXMLOnly
	default:
		ResponseTypePolicy = ContentTypePolicyJSONOrXML
	}
}

type encoder interface {
	Encode(interface{}) error
}

// Response wraps an http.ResponseWriter and handles content-type encoding.
type Response struct {
	ContentType string

	w http.ResponseWriter

	encoder
}

// NewResponse looks at the `Accept` header and returns an appropriate Response
// struct.
func NewResponse(w http.ResponseWriter, r *http.Request) Response {
	ct, enc := fromAccept(w, r.Header.Get("Accept"))
	return Response{
		ContentType: ct,
		encoder:     enc,
		w:           w,
	}
}

func fromAccept(w http.ResponseWriter, accept string) (string, encoder) {
	switch ResponseTypePolicy {
	case ContentTypePolicyJSONOnly:
		return ContentTypeJSON, json.NewEncoder(w)
	case ContentTypePolicyXMLOnly:
		return ContentTypeXML, xml.NewEncoder(w)
	default:
		if strings.Contains(accept, "xml") {
			return ContentTypeXML, xml.NewEncoder(w)
		}
		return ContentTypeJSON, json.NewEncoder(w)
	}
}

// WriteHeader sets the correct `Content-Type` header and writes a status
// code. All other headers should be set before calling this func.
func (res Response) WriteHeader(status int) {
	res.w.Header().Set("Content-Type", res.ContentType)
	res.w.WriteHeader(status)
}

// Send calls WriteHeader and serializes an interface using the determined
// `Content-Type`.
func (res Response) Send(status int, x interface{}) {
	res.w.WriteHeader(status)

	if err := res.Encode(x); err != nil {
		log.Warn(errors.Wrap(err, "unable to marshal response to content type"))
	}
}

// SendErr writes a standardized error struct. For 5XX errors, the message sent
// across the wire will be generic while the real error message is logged.
func (res Response) SendErr(status int, err error) {
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

	res.Send(status, struct {
		XMLName xml.Name `json:"-" xml:"error"`
		Error   error    `json:"error" xml:"message"`
	}{Error: err})
}
