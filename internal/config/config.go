// Package config отвечает за конфигурацию сервиса.
// Параметры читаются из флагов командной строки, переменные окружения имеют приоритет.
package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
)

// Config хранит настройки запуска сервиса.
type Config struct {
	// ServerAddress — адрес и порт, на котором слушает сервер.
	ServerAddress string
	// BaseURL — базовый URL для формирования коротких ссылок.
	BaseURL string
	// FileStoragePath — путь к файлу для хранения URL (если не используется БД).
	FileStoragePath string
	// DatabaseDSN — строка подключения к PostgreSQL.
	DatabaseDSN string
	// AuditFile — путь к файлу для записи аудит-событий.
	AuditFile string
	// AuditURL — внешний URL для отправки аудит-событий.
	AuditURL string
	// EnableHTTPS — включить HTTPS-режим сервера.
	EnableHTTPS bool
	// TrustedSubnet — доверенная подсеть
	TrustedSubnet string
}

type fileConfig struct {
	ServerAddress   *string `json:"server_address"`
	BaseURL         *string `json:"base_url"`
	FileStoragePath *string `json:"file_storage_path"`
	DatabaseDSN     *string `json:"database_dsn"`
	EnableHTTPS     *bool   `json:"enable_https"`
	TrustedSubnet   *string `json:"trusted_subnet"`
}

func loadFileConfig(path string) (*fileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, err
	}
	return &fc, nil
}

// SetupConfig разбирает флаги и переменные окружения, возвращает заполненный Config.
func SetupConfig() *Config {
	cfg := &Config{}

	var configFile string
	flag.StringVar(&configFile, "c", "", "Config file path")
	flag.StringVar(&configFile, "config", "", "Config file path")

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "short-url.json", "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "Audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "Audit URL")
	flag.BoolVar(&cfg.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "Trusted subnet")

	flag.Parse()

	if v, ok := os.LookupEnv("CONFIG"); ok {
		configFile = v
	}

	if configFile != "" {
		fc, err := loadFileConfig(configFile)
		if err == nil {
			setFlags := make(map[string]bool)
			flag.Visit(func(f *flag.Flag) {
				setFlags[f.Name] = true
			})

			if fc.ServerAddress != nil && !setFlags["a"] {
				cfg.ServerAddress = *fc.ServerAddress
			}
			if fc.BaseURL != nil && !setFlags["b"] {
				cfg.BaseURL = *fc.BaseURL
			}
			if fc.FileStoragePath != nil && !setFlags["f"] {
				cfg.FileStoragePath = *fc.FileStoragePath
			}
			if fc.DatabaseDSN != nil && !setFlags["d"] {
				cfg.DatabaseDSN = *fc.DatabaseDSN
			}
			if fc.EnableHTTPS != nil && !setFlags["s"] {
				cfg.EnableHTTPS = *fc.EnableHTTPS
			}
			if fc.TrustedSubnet != nil && !setFlags["t"] {
				cfg.TrustedSubnet = *fc.TrustedSubnet
			}
		}
	}

	if v, ok := os.LookupEnv("SERVER_ADDRESS"); ok {
		cfg.ServerAddress = v
	}
	if v, ok := os.LookupEnv("BASE_URL"); ok {
		cfg.BaseURL = v
	}
	if v, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		cfg.FileStoragePath = v
	}
	if v, ok := os.LookupEnv("DATABASE_DSN"); ok {
		cfg.DatabaseDSN = v
	}
	if v, ok := os.LookupEnv("AUDIT_FILE"); ok {
		cfg.AuditFile = v
	}
	if v, ok := os.LookupEnv("AUDIT_URL"); ok {
		cfg.AuditURL = v
	}
	if v, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		cfg.EnableHTTPS, _ = strconv.ParseBool(v)
	}
	if v, ok := os.LookupEnv("TRUSTED_SUBNET"); ok {
		cfg.TrustedSubnet = v
	}

	return cfg
}
