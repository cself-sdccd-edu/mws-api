/*
api/middleware.go
Authorization middleware for protected endpoints
Request logging middleware for event logging
*/
package api

import (
	"crypto/subtle"
	"net/http"

	"github.com/cself-sdccd-edu/mws-api/internal/config"
	mwslog "github.com/cself-sdccd-edu/mws-api/internal/log"
	"log"
	"time"
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

func requestMiddleware(logger mwslog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID := mwslog.NewRequestID()
		ctx := mwslog.WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)

		event := mwslog.LogEvent{
			RequestID: requestID,
			Event:     "request_start",
			ClientIP:  mwslog.ClientIP(r),
			UserAgent: r.UserAgent(),
			Method:    r.Method,
			Endpoint:  r.URL.Path,
		}

		if err := logger.Log(ctx, event); err != nil {
			log.Printf("request start logging failed: %v", err)
		}

		recorder := &statusRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)

		if recorder.statusCode == 0 {
			recorder.statusCode = http.StatusOK
		}

		event = mwslog.LogEvent{
			RequestID:  requestID,
			Event:      "request_complete",
			ClientIP:   mwslog.ClientIP(r),
			UserAgent:  r.UserAgent(),
			Method:     r.Method,
			Endpoint:   r.URL.Path,
			StatusCode: recorder.statusCode,
			DurationMs: int(duration.Milliseconds()),
			Message:    "request completed",
		}

		if err := logger.Log(ctx, event); err != nil {
			log.Printf("request completion logging failed: %v", err)
		}
	})
}
