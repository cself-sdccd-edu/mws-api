package main

import (
	"context"
	"encoding/json"
	"github.com/cself-sdccd-edu/mws-api/internal/api"
	"github.com/cself-sdccd-edu/mws-api/internal/cache"
	"github.com/cself-sdccd-edu/mws-api/internal/config"
	"github.com/cself-sdccd-edu/mws-api/internal/db"
	"github.com/cself-sdccd-edu/mws-api/internal/qas"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// first load configuration and exit if there's a failure
	configPath := os.Getenv("MWSAPI_CONFIG")
	if configPath == "" {
		configPath = "config/app.json"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}
	log.Printf("loading configuration from %s", configPath)
	safeConfig := cfg
	safeConfig.SQLPassword = ""
	safeConfig.QASPassword = ""
	safeConfig.AuthSecret = ""
	configJSON, _ := json.MarshalIndent(safeConfig, "", "    ")
	log.Printf("configuration:\n%s", configJSON)

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
	httpClient := &http.Client{Timeout: 30 * time.Second}
	qasClient := qas.NewHTTPClient(httpClient, cfg.QASDomain, cfg.QASSuffix, cfg.QueryParams, cfg.QASUser, cfg.QASPassword)
	cacheStore := cache.NewSQLServerStore(database)
	cacheService := cache.NewService(cacheStore, qasClient, log.Default(), time.Duration(cfg.CacheTime)*time.Second, time.Duration(cfg.MaxCacheTime)*time.Second, time.Duration(cfg.RefreshLeaseTime)*time.Second)

	server := api.NewServer(cfg, cacheService)
	log.Printf("mws-api starting on %s", cfg.Addr)
	log.Printf("environment: %s", cfg.SystemVersion)
	log.Printf("connected to SQL Server database %s", cfg.SQLDatabase)
	log.Fatal(server.ListenAndServe())

}
