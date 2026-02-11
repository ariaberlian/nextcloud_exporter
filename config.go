package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	// DefaultFetchInterval is the default minimum interval between API fetches (10 seconds)
	DefaultFetchInterval = 10 * time.Second

	// DefaultTimeout is the default HTTP client timeout
	DefaultTimeout = 5 * time.Second

	// DefaultListenAddr is the default address to listen on
	DefaultListenAddr = ":9205"
)

// Config holds all configuration for the exporter
type Config struct {
	BaseURL       string
	Token         string
	ListenAddr    string
	FetchInterval time.Duration
	Timeout       time.Duration
}

// LoadConfig loads configuration from command line flags and environment variables
func LoadConfig() *Config {
	// Command line flags
	baseURL := flag.String("url", "", "Nextcloud base URL (e.g., https://cloud.example.com)")
	token := flag.String("token", "", "NC-Token for authentication")
	listenAddr := flag.String("listen", "", "Address to listen on (default :9205)")
	fetchInterval := flag.Duration("fetch-interval", 0, "Minimum interval between API fetches to avoid rate limiting (default 30s)")
	timeout := flag.Duration("timeout", 0, "HTTP client timeout (default 10s)")
	flag.Parse()

	config := &Config{
		BaseURL:       *baseURL,
		Token:         *token,
		ListenAddr:    *listenAddr,
		FetchInterval: *fetchInterval,
		Timeout:       *timeout,
	}

	// Use environment variables as fallback
	if config.BaseURL == "" {
		config.BaseURL = getEnv("NEXTCLOUD_URL", "")
	}
	if config.Token == "" {
		config.Token = getEnv("NC_TOKEN", "")
	}
	if config.ListenAddr == "" {
		config.ListenAddr = getEnv("LISTEN_ADDR", DefaultListenAddr)
	}
	if config.FetchInterval == 0 {
		config.FetchInterval = getEnvDuration("FETCH_INTERVAL", DefaultFetchInterval)
	}
	if config.Timeout == 0 {
		config.Timeout = getEnvDuration("TIMEOUT", DefaultTimeout)
	}

	// Validate required parameters
	if config.BaseURL == "" {
		log.Fatal("Nextcloud URL is required. Set via -url flag or NEXTCLOUD_URL environment variable")
	}
	if config.Token == "" {
		log.Fatal("NC-Token is required. Set via -token flag or NC_TOKEN environment variable")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		// Try parsing as duration string (e.g., "30s", "1m")
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		// Try parsing as seconds (e.g., "30")
		if secs, err := strconv.Atoi(value); err == nil {
			return time.Duration(secs) * time.Second
		}
		log.Printf("Warning: invalid duration value for %s: %s, using default", key, value)
	}
	return defaultValue
}
