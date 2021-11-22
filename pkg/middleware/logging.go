package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/urfave/negroni"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// userAgent := r.Header.Get("User-Agent")
		// if strings.HasPrefix(userAgent, "Prometheus/") || strings.HasPrefix(userAgent, "kube-probe/") {
		// 	next.ServeHTTP(w, r)
		// 	return
		// }

		lrw := negroni.NewResponseWriter(w)
		start := time.Now()

		next.ServeHTTP(lrw, r)

		statusCode := lrw.Status()
		if shouldPrintToStderr(statusCode) {
			latency := time.Since(start).Seconds()
			request_id := r.Header.Get("X-Request-Id")

			log.Printf("request-id: %s, latency: %fs, status code: %d, %s %s", request_id, latency, statusCode, r.Method, r.URL)
		}
	})
}

func shouldPrintToStderr(statusCode int) bool {
	// 200 to 300 range
	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return false
	}

	// redirect to storage
	if statusCode == http.StatusTemporaryRedirect {
		return false
	}

	return true
}
