package web

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
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

	ErrUnsupportedMediaType = errors.New(http.StatusText(http.StatusUnsupportedMediaType))
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

// NewDecoder inspects the `Content-Type` header and either returns a
// http.Request.Body wrapped JSON or XML Decoder. If the Content-Type is
// not recognized it will return nil with `ErrUnsupportedMediaType`.
func NewDecoder(r *http.Request) (Decoder, error) {
	var d Decoder
	h := r.Header.Get("Content-Type")
	switch ResponseTypePolicy {
	case ContentTypePolicyJSONOnly:
		if !strings.Contains(h, "json") {
			return nil, ErrUnsupportedMediaType
		}
		d = json.NewDecoder(r.Body)
	case ContentTypePolicyXMLOnly:
		if !strings.Contains(h, "xml") {
			return nil, ErrUnsupportedMediaType
		}
		d = xml.NewDecoder(r.Body)
	default:
		if strings.Contains(h, "xml") {
			d = xml.NewDecoder(r.Body)
		} else if strings.Contains(h, "xml") {
			d = json.NewDecoder(r.Body)
		} else {
			return nil, ErrUnsupportedMediaType
		}
	}
	return d, nil
}

type Decoder interface {
	Decode(interface{}) error
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
	ct, enc := headerToEncoder(w, r.Header.Get("Accept"))
	return Response{
		ContentType: ct,
		encoder:     enc,
		w:           w,
	}
}

func headerToEncoder(w http.ResponseWriter, accept string) (string, encoder) {
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

	fmt.Println("err", err)
	res.Send(status, struct {
		XMLName xml.Name `json:"-" xml:"error"`
		Error   string   `json:"error" xml:"message"`
	}{Error: err.Error()})
}
