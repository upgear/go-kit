package web

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/upgear/go-kit/log"
)

// Logware logs requests.
func Logware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("new request", log.M{"method": r.Method, "path": r.URL.Path})
		next.ServeHTTP(w, r)
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
	case ContentTypePolicyJSON:
		setCT(ContentTypeJSON)
	case ContentTypePolicyXML:
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
		case ContentTypePolicyJSONOrXML, ContentTypePolicyJSON:
			return true
		}
	case "xml":
		switch GlobalContentTypePolicy {
		case ContentTypePolicyJSONOrXML, ContentTypePolicyXML:
			return true
		}
	}

	return false
}
