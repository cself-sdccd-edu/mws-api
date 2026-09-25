/*
api/auth.go
Authorization middleware for protected endpoints
*/
package api

import (
	"crypto/subtle"
	"net/http"

	"github.com/cself-sdccd-edu/mws-api/internal/config"
)

func authMiddleware(cfg config.Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get(cfg.AuthHeader)

		if subtle.ConstantTimeCompare(
			[]byte(authorization),
			[]byte(cfg.AuthSecret),
		) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
