package api

import (
	"github.com/cself-sdccd-edu/mws-api/internal/cache"
	"github.com/cself-sdccd-edu/mws-api/internal/config"
	mwslog "github.com/cself-sdccd-edu/mws-api/internal/log"
	"net/http"
	"time"
)

type Server struct {
	config config.Config
	cache  *cache.Service
	logger mwslog.Logger
}

func NewServer(cfg config.Config, cache *cache.Service, logger mwslog.Logger) *http.Server {
	server := Server{
		config: cfg,
		cache:  cache,
		logger: logger,
	}
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           server.routes(),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", healthHandler)

	mux.Handle(
		"GET /schedule",
		requestMiddleware(s.logger,
			authMiddleware(
				s.config,
				http.HandlerFunc(s.endpointHandler),
			),
		),
	)

	// backwards compatible routes (may be removed later)
	mux.Handle(
		"GET /{$}",
		requestMiddleware(s.logger,
			authMiddleware(
				s.config,
				http.HandlerFunc(s.scheduleHandler),
			),
		),
	)
	mux.Handle(
		"GET /index.cfm",
		requestMiddleware(s.logger,
			authMiddleware(
				s.config,
				http.HandlerFunc(s.scheduleHandler),
			),
		),
	)

	return mux
}
