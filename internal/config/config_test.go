package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestSetupConfig_Defaults(t *testing.T) {
	resetFlags()
	os.Clearenv()

	cfg := SetupConfig()

	assert.Equal(t, "localhost:8080", cfg.ServerAddress)
	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
	assert.Equal(t, "short-url.json", cfg.FileStoragePath)
	assert.Equal(t, "", cfg.DatabaseDSN)
	assert.Equal(t, "", cfg.AuditFile)
	assert.Equal(t, "", cfg.AuditURL)
}

func TestSetupConfig_EnvOverrides(t *testing.T) {
	resetFlags()
	os.Clearenv()
	_ = os.Setenv("SERVER_ADDRESS", "0.0.0.0:9090")
	_ = os.Setenv("BASE_URL", "http://example.com")
	_ = os.Setenv("FILE_STORAGE_PATH", "/tmp/urls.json")
	_ = os.Setenv("DATABASE_DSN", "postgres://localhost/test")
	_ = os.Setenv("AUDIT_FILE", "/tmp/audit.log")
	_ = os.Setenv("AUDIT_URL", "http://audit.example.com")
	defer os.Clearenv()

	cfg := SetupConfig()

	assert.Equal(t, "0.0.0.0:9090", cfg.ServerAddress)
	assert.Equal(t, "http://example.com", cfg.BaseURL)
	assert.Equal(t, "/tmp/urls.json", cfg.FileStoragePath)
	assert.Equal(t, "postgres://localhost/test", cfg.DatabaseDSN)
	assert.Equal(t, "/tmp/audit.log", cfg.AuditFile)
	assert.Equal(t, "http://audit.example.com", cfg.AuditURL)
}
