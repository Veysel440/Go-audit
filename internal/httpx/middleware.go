package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/Veysel440/go-audit/pkg/rate"
)

func withLogging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := rid()
		start := time.Now()
		ww := &wrap{ResponseWriter: w, code: 200}
		logger.Info("req_start", "id", id, "m", r.Method, "p", r.URL.Path)
		next.ServeHTTP(ww, r)
		logger.Info("req_end", "id", id, "status", ww.code, "ms", time.Since(start).Milliseconds())
	})
}

func withRecover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				writeErr(w, http.StatusInternalServerError, "internal_error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func withRate(l *rate.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if l != nil && !l.Allow(rate.IP(r.RemoteAddr)) {
				writeErr(w, http.StatusTooManyRequests, "rate_limited")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type wrap struct {
	http.ResponseWriter
	code int
}

func (w *wrap) WriteHeader(c int) { w.code = c; w.ResponseWriter.WriteHeader(c) }

func rid() string { b := make([]byte, 8); _, _ = rand.Read(b); return hex.EncodeToString(b) }
