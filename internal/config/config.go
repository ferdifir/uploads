package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	APIKey         string `json:"api_key"`
	UIUsername     string `json:"ui_username"`
	UIPasswordHash string `json:"ui_password_hash"`
	DataDir        string `json:"data_dir"`
	DBPath         string `json:"db_path"`
	Port           int    `json:"port"`
}

func (c Config) ServerAddr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set default values if not provided
	if config.DataDir == "" {
		config.DataDir = "data"
	}
	if config.DBPath == "" {
		config.DBPath = "uploads.db"
	}
	if config.Port == 0 {
		config.Port = 8080
	}

	return &config, nil
}
