package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	Addr             string            `json:"addr"`
	SystemVersion    string            `json:"system_version"`
	ServerNumber     string            `json:"server_number"`
	SQLServer        string            `json:"sql_server"`
	SQLDatabase      string            `json:"sql_database"`
	CacheTime        int               `json:"cache_time"`
	RefreshLeaseTime int               `json:"refresh_lease_time"`
	QASDomain        string            `json:"qas_domain"`
	QASSuffix        string            `json:"qas_url_suffix"`
	QueryParams      map[string]string `json:"query_params"`
	Queries          map[string]string `json:"queries"`
	AuthHeader       string            `json:"auth_header"`
	CORSOrigins      []string          `json:"cors_origins"`

	SQLUser     string `json:"-"`
	SQLPassword string `json:"-"`
	QASUser     string `json:"-"`
	QASPassword string `json:"-"`
	AuthSecret  string `json:"-"`
}

func Load(path string) (Config, error) {
	if path == "" {
		path = os.Getenv("MWSAPI_CONFIG")
	}

	if path == "" {
		path = "config/app.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}

	cfg.SQLUser = os.Getenv("MWSAPI_SQL_USER")
	cfg.SQLPassword = os.Getenv("MWSAPI_SQL_PASSWORD")
	cfg.QASUser = os.Getenv("MWSAPI_QAS_USER")
	cfg.QASPassword = os.Getenv("MWSAPI_QAS_PASSWORD")
	cfg.AuthSecret = os.Getenv("MWSAPI_AUTH_SECRET")

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (cfg Config) Validate() error {
	required := map[string]string{
		"sql_server":     cfg.SQLServer,
		"sql_database":   cfg.SQLDatabase,
		"qas_domain":     cfg.QASDomain,
		"qas_url_suffix": cfg.QASSuffix,
		"auth_header":    cfg.AuthHeader,
		"sql_user":       cfg.SQLUser,
		"sql_password":   cfg.SQLPassword,
		"qas_user":       cfg.QASUser,
		"qas_password":   cfg.QASPassword,
		"auth_secret":    cfg.AuthSecret,
	}

	for name, value := range required {
		if value == "" {
			return fmt.Errorf("required configuration value %q is not set", name)
		}
	}

	if cfg.CacheTime <= 0 {
		return errors.New("cache_time must be greater than zero")
	}

	if cfg.RefreshLeaseTime <= 0 {
		return errors.New("refresh_lease_time must be greater than zero")
	}

	if len(cfg.Queries) == 0 {
		return errors.New("queries must contain at least one query")
	}

	return nil
}
