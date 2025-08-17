package httpx

import (
	"net/http"
	"strings"
)

type APIKeyAuth struct {
	keys map[string]struct{}
}

func NewAPIKeyAuth(csv string) *APIKeyAuth {
	m := map[string]struct{}{}
	for _, k := range strings.Split(csv, ",") {
		k = strings.TrimSpace(k)
		if k != "" {
			m[k] = struct{}{}
		}
	}
	return &APIKeyAuth{keys: m}
}

func (a *APIKeyAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(a.keys) == 0 {
			next.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get("X-Api-Key")
		if _, ok := a.keys[key]; !ok {
			writeErr(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}
