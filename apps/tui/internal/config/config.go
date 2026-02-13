package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	defaultAPIBaseURL  = "http://localhost:8080"
	defaultAppDataName = "libra-link-tui"
	defaultHTTPTimeout = 15 * time.Second
	defaultSyncEvery   = 10 * time.Second
)

type Config struct {
	APIBaseURL    string
	DataDir       string
	DBPath        string
	SessionPath   string
	HTTPTimeout   time.Duration
	SyncInterval  time.Duration
	SyncBatchSize int
}

func Load() (*Config, error) {
	apiBaseURL := envOrDefault("LIBRA_TUI_API_BASE_URL", defaultAPIBaseURL)

	dataDir, err := resolveDataDir()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}

	httpTimeout := durationFromEnv("LIBRA_TUI_HTTP_TIMEOUT_SECONDS", defaultHTTPTimeout)
	syncInterval := durationFromEnv("LIBRA_TUI_SYNC_INTERVAL_SECONDS", defaultSyncEvery)
	syncBatchSize := intFromEnv("LIBRA_TUI_SYNC_BATCH_SIZE", 25)
	if syncBatchSize < 1 {
		syncBatchSize = 1
	}

	cfg := &Config{
		APIBaseURL:    apiBaseURL,
		DataDir:       dataDir,
		DBPath:        filepath.Join(dataDir, "libra-link.db"),
		SessionPath:   filepath.Join(dataDir, "session.json"),
		HTTPTimeout:   httpTimeout,
		SyncInterval:  syncInterval,
		SyncBatchSize: syncBatchSize,
	}
	return cfg, nil
}

func resolveDataDir() (string, error) {
	if custom := os.Getenv("LIBRA_TUI_DATA_DIR"); custom != "" {
		return custom, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if home == "" {
		return "", errors.New("home directory is empty")
	}

	return filepath.Join(home, ".local", "share", defaultAppDataName), nil
}

func envOrDefault(key, fallback string) string {
	if raw := os.Getenv(key); raw != "" {
		return raw
	}
	return fallback
}

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

func intFromEnv(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
