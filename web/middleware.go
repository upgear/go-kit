package web

import (
	"net/http"

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

func Logger(next http.Handler) http.Handler {
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
