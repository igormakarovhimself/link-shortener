package config

import "flag"

type Config struct {
	ServerAddress string
	BaseURL       string
}

func ParseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "Server address")
	flag.StringVar(&cfg.BaseURL, "b", "http://localhost:8080", "Base URL")

	flag.Parse()

	return cfg
}
