package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/cself-sdccd-edu/mws-api/internal/api"
	"github.com/cself-sdccd-edu/mws-api/internal/cache"
	"github.com/cself-sdccd-edu/mws-api/internal/config"
	"github.com/cself-sdccd-edu/mws-api/internal/db"
	mwslog "github.com/cself-sdccd-edu/mws-api/internal/log"
	"github.com/cself-sdccd-edu/mws-api/internal/qas"
	"github.com/cself-sdccd-edu/mws-api/internal/version"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultConfigPath = "/etc/mwsapi/config.json"

func main() {
	// first load configuration and exit if there's a failure
	configPath := flag.String("config", defaultConfigPath, "path to application configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}
	log.Printf("loading configuration from %s", *configPath)
	log.Printf("environment: %s", cfg.SystemVersion)
	log.Printf("version: %s %s %s", version.Version, version.Commit, version.BuildDate)

	// create context for our sevices
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// open a database object and exit if there's a failure
	database, err := db.Open(ctx, cfg.SQLServer, cfg.SQLDatabase, cfg.SQLUser, cfg.SQLPassword, cfg.TrustServerCertificate)
	if err != nil {
		log.Printf("database error: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	// init services
	appLogger := mwslog.NewSQLLogger(database, cfg.ServerNumber, version.Version)
	// qas client, timeout value matches the refresh timeout
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.RefreshTimeout) * time.Second,
	}
	qasClient := qas.NewHTTPClient(
		httpClient,
		cfg.QASDomain,
		cfg.QASSuffix,
		cfg.Endpoints,
		cfg.QASUser,
		cfg.QASPassword,
	)
	cacheStore := cache.NewSQLServerStore(database)
	cacheService := cache.NewService(
		cacheStore,
		qasClient,
		log.Default(),
		appLogger,
		time.Duration(cfg.CacheTime)*time.Second,
		time.Duration(cfg.MaxCacheTime)*time.Second,
		time.Duration(cfg.RefreshLeaseTime)*time.Second,
		time.Duration(cfg.RefreshTimeout)*time.Second,
	)

	httpServer := api.NewServer(cfg, cacheService, appLogger)
	log.Printf("mws-api starting on %s", cfg.Addr)
	log.Printf("connected to SQL Server database %s", cfg.SQLDatabase)

	// configure the listen requirements, start the server, and listen for shutdown signals
	ln, err := net.Listen("tcp4", cfg.Addr)
	if err != nil {
		log.Fatal(err)
	}
	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- httpServer.Serve(ln)
	}()

	// warm up the cache immediately if the config says to
	if cfg.CacheWarmup {
		go func() {
			warmupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			warmCache(warmupCtx, cacheService, cfg)
		}()
	}

	shutdownSignal := make(chan os.Signal, 1)

	signal.Notify(
		shutdownSignal,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}

	case sig := <-shutdownSignal:
		log.Printf("shutdown signal received: %v", sig)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)

		if err := httpServer.Close(); err != nil {
			log.Printf("forced shutdown failed: %v", err)
		}
	}

	log.Println("shutdown complete")
}

func startupTerms(now time.Time) []string {
	year := now.Year()
	termYear := fmt.Sprintf("%d%02d", year/1000, year%100)

	switch now.Month() {
	case time.January, time.February, time.March:
		return []string{termYear + "3", termYear + "5"}

	case time.April, time.May, time.June, time.July, time.August, time.September:
		return []string{termYear + "5", termYear + "7"}

	default:
		nextYear := year + 1
		nextTermYear := fmt.Sprintf("%d%02d", nextYear/1000, nextYear%100)
		return []string{nextTermYear + "3"}
	}
}

func warmCache(ctx context.Context, cacheService *cache.Service, cfg config.Config) {
	terms := startupTerms(time.Now())
	careers := []string{"ugrd", "ce"}

	endpointConfig, ok := cfg.Endpoints["schedule"]
	if !ok {
		log.Printf("cache warm-up skipped: schedule endpoint is not configured")
		return
	}

	log.Printf("warming cache for terms: %v", terms)

	for _, term := range terms {
		for _, career := range careers {
			queryName, ok := endpointConfig.Queries[career]
			if !ok {
				log.Printf("cache warm-up skipped %s/%s: query not configured", term, career)
				continue
			}

			key := fmt.Sprintf("schedule:%s:%s", term, career)

			log.Printf("warming cache: %s", key)

			_, err := cacheService.Get(ctx, key, cache.RefreshRequest{
				EndpointName: "schedule",
				QueryName:    queryName,
				Term:         term,
			})
			if err != nil {
				log.Printf("cache warm-up failed for %s: %v", key, err)
				continue
			}

			log.Printf("cache warm-up completed: %s", key)
		}
	}
}
