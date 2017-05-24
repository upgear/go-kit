package web

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/upgear/go-kit/log"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// Logware logs requests after nested handlers are complete.
func Logware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default to 200, like the handler will
		sw := statusWriter{w, http.StatusOK}

		// Call the next handler and record any calls made to
		// ResponseWriter.WriteHeader
		next.ServeHTTP(&sw, r)

		kvs := log.M{"method": r.Method, "path": r.URL.Path, "status": sw.status}
		const msg = "served request"
		if sw.status >= 500 {
			log.Error(msg, kvs)
		} else if sw.status >= 400 {
			log.Warn(msg, kvs)
		} else {
			log.Info(msg, kvs)
		}

		return
	})
}

var errUnsupportedMediaType = errors.New(http.StatusText(http.StatusUnsupportedMediaType))

// Contentware sets the response `Content-Type` header.
// It responds with an "Unsupported Media Type" error when conflicting request
// content types are provided.
func Contentware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCT := r.Header.Get("Content-Type")

		if ct := "xml"; strings.Contains(reqCT, ct) && supported(ct) {
			Error(w, errors.Wrapf(errUnsupportedMediaType, "Content-Type %s not supported", strings.ToUpper(ct)), http.StatusUnsupportedMediaType)
			return
		}
		if ct := "json"; strings.Contains(reqCT, ct) && supported(ct) {
			Error(w, errors.Wrapf(errUnsupportedMediaType, "Content-Type %s not supported", strings.ToUpper(ct)), http.StatusUnsupportedMediaType)
			return
		}

		setCTFromAccept(w, r)

		next.ServeHTTP(w, r)
	})
}

func setCTFromAccept(w http.ResponseWriter, r *http.Request) {
	setCT := func(ct ContentType) {
		w.Header().Set("Content-Type", string(ct))
	}

	switch GlobalContentTypePolicy {
	case ContentTypePolicyJSONOnly:
		setCT(ContentTypeJSON)
	case ContentTypePolicyXMLOnly:
		setCT(ContentTypeXML)
	default:
		if strings.Contains(r.Header.Get("Accept"), "xml") {
			setCT(ContentTypeXML)
		} else {
			setCT(ContentTypeJSON)
		}
	}
}

func supported(ct string) bool {
	switch ct {
	case "json":
		switch GlobalContentTypePolicy {
		case ContentTypePolicyJSONOrXML, ContentTypePolicyJSONOnly:
			return true
		}
	case "xml":
		switch GlobalContentTypePolicy {
		case ContentTypePolicyJSONOrXML, ContentTypePolicyXMLOnly:
			return true
		}
	}

	return false
}
