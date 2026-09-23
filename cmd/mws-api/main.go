package main

import (
	"log"
	"os"

	"github.com/cself-sdccd-edu/mws-api/internal/config"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}

	log.Printf("mws-api starting on %s", cfg.Addr)
	log.Printf("environment: %s", cfg.SystemVersion)
}
