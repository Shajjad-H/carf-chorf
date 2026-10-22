package log_middleware

import (
	"net/http"

	"carf-chorf.com/app/pkg/simple_logger"
)

func LogHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			simple_logger.Debugf(true, "%s: %s", k, v)
		}
		next.ServeHTTP(w, r)
	})
}
