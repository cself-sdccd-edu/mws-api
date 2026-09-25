package main

import (
	"context"
	"log"
	"os"
	"time"
	"github.com/cself-sdccd-edu/mws-api/internal/config"
	"github.com/cself-sdccd-edu/mws-api/internal/db"
)

func main() {
	configPath := os.Getenv("MWSAPI_CONFIG")
	if configPath == "" {
		configPath = "config/app.json"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}

	//ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.Open(ctx, cfg.SQLServer, cfg.SQLDatabase, cfg.SQLUser, cfg.SQLPassword, cfg.TrustServerCertificate)
	if err != nil {
		log.Printf("database error: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	log.Printf("mws-api starting on %s", cfg.Addr)
	log.Printf("environment: %s", cfg.SystemVersion)
	log.Printf("connected to SQL Server database %s", cfg.SQLDatabase)
}
