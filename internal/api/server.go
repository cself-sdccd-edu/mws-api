package api

import (
	"github.com/cself-sdccd-edu/mws-api/internal/cache"
	"github.com/cself-sdccd-edu/mws-api/internal/config"
	"net/http"
)

type Server struct {
	config config.Config
	cache  *cache.Service
}

func NewServer(cfg config.Config, cache *cache.Service) *http.Server {
	server := Server{
		config: cfg,
		cache:  cache,
	}
	return &http.Server{
		Addr:    cfg.Addr,
		Handler: server.routes(),
	}
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", healthHandler)

	mux.HandleFunc("/schedule", scheduleHandler)

	return mux
}
