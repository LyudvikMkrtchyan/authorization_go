package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DBHost            string `json:"db_host"`
	DBPort            int    `json:"db_port"`
	DBName            string `json:"db_name"`
	DBUser            string `json:"db_user"`
	DBPassword        string `json:"db_password"`
	JWTSecret         string `json:"jwt_secret"`
	JWTTTLSeconds     int    `json:"jwt_ttl_seconds"`
	RefreshTTLSeconds int    `json:"refresh_ttl_seconds"`
	TokenMapMax       int    `json:"token_map_max"`
	WarningLogDir     string `json:"warning_log_dir"`
}

type ConfigLoader interface {
	Load(path string) (*Config, error)
}

type JSONConfigLoader struct{}

func (l *JSONConfigLoader) Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot open %q: %w", path, err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("config: cannot parse %q: %w", path, err)
	}
	return &cfg, nil
}
