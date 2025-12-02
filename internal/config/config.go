package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func SetupConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "short-url.json", "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN")

	flag.Parse()

	if envServerAddress := os.Getenv("SERVER_ADDRESS"); envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		cfg.BaseURL = envBaseURL
	}

	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		cfg.FileStoragePath = envFilePath
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

	return cfg
}
