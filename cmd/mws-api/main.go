package main

import (
	"context"
	//"fmt"
	"github.com/cself-sdccd-edu/mws-api/internal/cache"
	"github.com/cself-sdccd-edu/mws-api/internal/config"
	"github.com/cself-sdccd-edu/mws-api/internal/db"
	"github.com/cself-sdccd-edu/mws-api/internal/qas"
	"github.com/cself-sdccd-edu/mws-api/internal/api"
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
	qasClient := qas.NewHTTPClient(http.DefaultClient, cfg.QASDomain, cfg.QASSuffix, cfg.QueryParams, cfg.QASUser, cfg.QASPassword)
	cacheStore := cache.NewSQLServerStore(database)
	cacheService := cache.NewService(cacheStore, qasClient, log.Default(), time.Duration(cfg.CacheTime)*time.Second, time.Duration(cfg.RefreshLeaseTime)*time.Second)

	server := api.NewServer(cfg,cacheService)
	log.Printf("mws-api starting on %s", cfg.Addr)
	log.Printf("environment: %s", cfg.SystemVersion)
	log.Printf("connected to SQL Server database %s", cfg.SQLDatabase)
	log.Fatal(server.ListenAndServe())

	/* this was temporary verification of database and cache features. it will be removed as the application grows
	store := cache.NewSQLServerStore(database)

	entry, err := store.Get(ctx, "test:1")
	if err != nil {
		log.Printf("cache error: %v", err)
		os.Exit(1)
	}

	if entry == nil {
		log.Printf("cache entry does not exist")
	} else {
		log.Printf("cache entry exists: updated=%s data=%d bytes", entry.UpdatedAt, len(entry.Data))
	}

	claimed, err := store.TryStartRefresh(ctx, "test:1", 120*time.Second)
	if err != nil {
		log.Printf("cache refresh error: %v", err)
		os.Exit(1)
	}
	log.Printf("cache refresh claimed: %t", claimed)

	if err := store.FailRefresh(ctx, "test:1", fmt.Errorf("test QAS failure")); err != nil {
		log.Printf("cache failure error: %v", err)
		os.Exit(1)
	}
	log.Printf("cache refresh failure recorded")
	*/
	/*
	   data := []byte(`{"test":"Cache_Save"}`)

	   	if err := store.Save(ctx, "test:1", data); err != nil {
	   		log.Printf("cache save error: %v", err)
	   		os.Exit(1)
	   	}

	   log.Printf("cache entry saved: %d bytes", len(data))
	*/
}
