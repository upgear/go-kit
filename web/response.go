package web

import (
	"encoding/json"
	"encoding/xml"
	"io"
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

type newEncoderFunc func(io.Reader) encoder

type encoder interface {
	Encode(interface{}) error
}

type Response struct {
	ContentType string

	w http.ResponseWriter

	encoder
}

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

func (res Response) WriteHeader(status int) {
	res.w.Header().Set("Content-Type", res.ContentType)
	res.w.WriteHeader(status)
}

func (res Response) Send(status int, x interface{}) {
	res.w.Header().Set("Content-Type", res.ContentType)
	res.w.WriteHeader(status)

	if err := res.Encode(x); err != nil {
		log.Warn(errors.Wrap(err, "unable to marshal response to content type"))
	}
}

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
