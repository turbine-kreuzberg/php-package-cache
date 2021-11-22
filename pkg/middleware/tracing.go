package middleware

import (
	"net/http"
	"unsafe"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// https://github.com/yurishkuro/opentracing-tutorial

func InitTraceContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		// name := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		name := r.Method
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		clientContext, _ := tracer.Extract(opentracing.HTTPHeaders, carrier)

		span := tracer.StartSpan(name, ext.RPCServerOption(clientContext))
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(r.Context(), span)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestID(randInt63 func() int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ID := r.Header.Get("X-Request-Id")
		if ID == "" {
			ID = randString(randInt63, 32)
		}

		span := opentracing.SpanFromContext(r.Context())
		span.SetTag("request-id", ID)

		r.Header.Set("X-Request-Id", ID)
		w.Header().Set("X-Request-Id", ID)

		next.ServeHTTP(w, r)
	})
}

// from https://stackoverflow.com/a/31832326
const (
	letterBytes   = "0123456789abcde"
	letterIdxBits = 4                    // 4 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randString(r func() int64, n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, r(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
