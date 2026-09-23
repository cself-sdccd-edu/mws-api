package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr             string
	SQLServer        string
	SQLDatabase      string
	SQLUser          string
	SQLPassword      string
	QASDomain        string
	QASSuffix        string
	QASUser          string
	QASPassword      string
	CacheTTLSeconds  int
	RefreshLeaseSecs int
	AuthHeader       string
	AuthSecret       string
	ServerNumber     string
	SystemVersion    string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:             getEnv("MWSAPI_ADDR", ":8080"),
		SQLServer:        os.Getenv("MWSAPI_SQL_SERVER"),
		SQLDatabase:      getEnv("MWSAPI_SQL_DATABASE", "MWSAPI"),
		SQLUser:          os.Getenv("MWSAPI_SQL_USER"),
		SQLPassword:      os.Getenv("MWSAPI_SQL_PASSWORD"),
		QASDomain:        os.Getenv("MWSAPI_QAS_DOMAIN"),
		QASSuffix:        os.Getenv("MWSAPI_QAS_SUFFIX"),
		QASUser:          os.Getenv("MWSAPI_QAS_USER"),
		QASPassword:      os.Getenv("MWSAPI_QAS_PASSWORD"),
		CacheTTLSeconds:  getEnvInt("MWSAPI_CACHE_TTL_SECONDS", 600),
		RefreshLeaseSecs: getEnvInt("MWSAPI_REFRESH_LEASE_SECONDS", 120),
		AuthHeader:       getEnv("MWSAPI_AUTH_HEADER", "X-NDnerBYETrTEHL6F"),
		AuthSecret:       os.Getenv("MWSAPI_AUTH_SECRET"),
		ServerNumber:     os.Getenv("MWSAPI_SERVER_NUMBER"),
		SystemVersion:    getEnv("MWSAPI_SYSTEM_VERSION", "development"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	required := map[string]string{
		"MWSAPI_SQL_SERVER":   cfg.SQLServer,
		"MWSAPI_QAS_DOMAIN":   cfg.QASDomain,
		"MWSAPI_QAS_SUFFIX":   cfg.QASSuffix,
		"MWSAPI_QAS_USER":     cfg.QASUser,
		"MWSAPI_QAS_PASSWORD": cfg.QASPassword,
		"MWSAPI_AUTH_SECRET":  cfg.AuthSecret,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required environment variable %s is not set", name)
		}
	}

	if cfg.CacheTTLSeconds <= 0 {
		return errors.New("MWSAPI_CACHE_TTL_SECONDS must be greater than zero")
	}

	if cfg.RefreshLeaseSecs <= 0 {
		return errors.New("MWSAPI_REFRESH_LEASE_SECONDS must be greater than zero")
	}

	return nil
}

func getEnv(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return result
}
