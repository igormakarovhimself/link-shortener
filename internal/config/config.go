// Package config отвечает за конфигурацию сервиса.
// Параметры читаются из флагов командной строки, переменные окружения имеют приоритет.
package config

import (
	"flag"
	"os"
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
}

// SetupConfig разбирает флаги и переменные окружения, возвращает заполненный Config.
func SetupConfig() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.StringVar(&cfg.FileStoragePath, "f", "short-url.json", "File storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "Database DSN")
	flag.StringVar(&cfg.AuditFile, "audit-file", "", "Audit file path")
	flag.StringVar(&cfg.AuditURL, "audit-url", "", "Audit URL")

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

	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		cfg.AuditURL = envAuditURL
	}

	return cfg
}
