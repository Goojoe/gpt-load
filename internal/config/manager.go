// Package config provides configuration management for the application
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"gpt-load/internal/errors"
	"gpt-load/pkg/types"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Constants represents configuration constants
type Constants struct {
	MinPort               int
	MaxPort               int
	MinTimeout            int
	DefaultTimeout        int
	DefaultMaxSockets     int
	DefaultMaxFreeSockets int
}

// DefaultConstants holds default configuration values
var DefaultConstants = Constants{
	MinPort:               1,
	MaxPort:               65535,
	MinTimeout:            1,
	DefaultTimeout:        30,
	DefaultMaxSockets:     50,
	DefaultMaxFreeSockets: 10,
}

// Manager implements the ConfigManager interface
type Manager struct {
	config            *Config
	roundRobinCounter uint64
}

// Config represents the application configuration
type Config struct {
	Server      types.ServerConfig      `json:"server"`
	Keys        types.KeysConfig        `json:"keys"`
	OpenAI      types.OpenAIConfig      `json:"openai"`
	Auth        types.AuthConfig        `json:"auth"`
	CORS        types.CORSConfig        `json:"cors"`
	Performance types.PerformanceConfig `json:"performance"`
	Log         types.LogConfig         `json:"log"`
}

// NewManager creates a new configuration manager
func NewManager() (types.ConfigManager, error) {
	// Try to load .env file
	if err := godotenv.Load(); err != nil {
		logrus.Info("Info: Create .env file to support environment variable configuration")
	}

	config := &Config{
		Server: types.ServerConfig{
			Port:                    parseInteger(os.Getenv("PORT"), 7860),
			Host:                    getEnvOrDefault("HOST", "0.0.0.0"),
			ReadTimeout:             parseInteger(os.Getenv("SERVER_READ_TIMEOUT"), 120),
			WriteTimeout:            parseInteger(os.Getenv("SERVER_WRITE_TIMEOUT"), 1800),
			IdleTimeout:             parseInteger(os.Getenv("SERVER_IDLE_TIMEOUT"), 120),
			GracefulShutdownTimeout: parseInteger(os.Getenv("SERVER_GRACEFUL_SHUTDOWN_TIMEOUT"), 60),
		},
		Keys: types.KeysConfig{
			APIKeys:            parseArray(os.Getenv("API_KEYS"), []string{}),
			StartIndex:         parseInteger(os.Getenv("START_INDEX"), 0),
			BlacklistThreshold: parseInteger(os.Getenv("BLACKLIST_THRESHOLD"), 1),
			MaxRetries:         parseInteger(os.Getenv("MAX_RETRIES"), 3),
		},
		OpenAI: types.OpenAIConfig{
			BaseURLs:        parseArray(os.Getenv("OPENAI_BASE_URL"), []string{"https://api.openai.com"}),
			RequestTimeout:  parseInteger(os.Getenv("REQUEST_TIMEOUT"), DefaultConstants.DefaultTimeout),
			ResponseTimeout: parseInteger(os.Getenv("RESPONSE_TIMEOUT"), 30),
			IdleConnTimeout: parseInteger(os.Getenv("IDLE_CONN_TIMEOUT"), 120),
		},
		Auth: types.AuthConfig{
			Key:     os.Getenv("AUTH_KEY"),
			Enabled: os.Getenv("AUTH_KEY") != "",
		},
		CORS: types.CORSConfig{
			Enabled:          parseBoolean(os.Getenv("ENABLE_CORS"), true),
			AllowedOrigins:   parseArray(os.Getenv("ALLOWED_ORIGINS"), []string{"*"}),
			AllowedMethods:   parseArray(os.Getenv("ALLOWED_METHODS"), []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			AllowedHeaders:   parseArray(os.Getenv("ALLOWED_HEADERS"), []string{"*"}),
			AllowCredentials: parseBoolean(os.Getenv("ALLOW_CREDENTIALS"), false),
		},
		Performance: types.PerformanceConfig{
			MaxConcurrentRequests: parseInteger(os.Getenv("MAX_CONCURRENT_REQUESTS"), 100),
			EnableGzip:            parseBoolean(os.Getenv("ENABLE_GZIP"), true),
		},
		Log: types.LogConfig{
			Level:         getEnvOrDefault("LOG_LEVEL", "info"),
			Format:        getEnvOrDefault("LOG_FORMAT", "text"),
			EnableFile:    parseBoolean(os.Getenv("LOG_ENABLE_FILE"), false),
			FilePath:      getEnvOrDefault("LOG_FILE_PATH", "logs/app.log"),
			EnableRequest: parseBoolean(os.Getenv("LOG_ENABLE_REQUEST"), true),
		},
	}

	manager := &Manager{config: config}

	// Validate configuration
	if err := manager.Validate(); err != nil {
		return nil, err
	}

	return manager, nil
}

// GetServerConfig returns server configuration
func (m *Manager) GetServerConfig() types.ServerConfig {
	return m.config.Server
}

// GetKeysConfig returns keys configuration
func (m *Manager) GetKeysConfig() types.KeysConfig {
	return m.config.Keys
}

// GetOpenAIConfig returns OpenAI configuration
func (m *Manager) GetOpenAIConfig() types.OpenAIConfig {
	config := m.config.OpenAI
	if len(config.BaseURLs) > 1 {
		// Use atomic counter for thread-safe round-robin
		index := atomic.AddUint64(&m.roundRobinCounter, 1) - 1
		config.BaseURL = config.BaseURLs[index%uint64(len(config.BaseURLs))]
	} else if len(config.BaseURLs) == 1 {
		config.BaseURL = config.BaseURLs[0]
	}
	return config
}

// GetAuthConfig returns authentication configuration
func (m *Manager) GetAuthConfig() types.AuthConfig {
	return m.config.Auth
}

// GetCORSConfig returns CORS configuration
func (m *Manager) GetCORSConfig() types.CORSConfig {
	return m.config.CORS
}

// GetPerformanceConfig returns performance configuration
func (m *Manager) GetPerformanceConfig() types.PerformanceConfig {
	return m.config.Performance
}

// GetLogConfig returns logging configuration
func (m *Manager) GetLogConfig() types.LogConfig {
	return m.config.Log
}

// Validate validates the configuration
func (m *Manager) Validate() error {
	var validationErrors []string

	// Validate port
	if m.config.Server.Port < DefaultConstants.MinPort || m.config.Server.Port > DefaultConstants.MaxPort {
		validationErrors = append(validationErrors, fmt.Sprintf("port must be between %d-%d", DefaultConstants.MinPort, DefaultConstants.MaxPort))
	}

	// Validate start index
	if m.config.Keys.StartIndex < 0 {
		validationErrors = append(validationErrors, "start index cannot be less than 0")
	}

	// Validate blacklist threshold
	if m.config.Keys.BlacklistThreshold < 1 {
		validationErrors = append(validationErrors, "blacklist threshold cannot be less than 1")
	}

	// Validate timeout
	if m.config.OpenAI.RequestTimeout < DefaultConstants.MinTimeout {
		validationErrors = append(validationErrors, fmt.Sprintf("request timeout cannot be less than %ds", DefaultConstants.MinTimeout))
	}

	// Validate upstream URL format
	if len(m.config.OpenAI.BaseURLs) == 0 {
		validationErrors = append(validationErrors, "at least one upstream API URL is required")
	}
	for _, baseURL := range m.config.OpenAI.BaseURLs {
		if _, err := url.Parse(baseURL); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("invalid upstream API URL format: %s", baseURL))
		}
	}

	// Validate performance configuration
	if m.config.Performance.MaxConcurrentRequests < 1 {
		validationErrors = append(validationErrors, "max concurrent requests cannot be less than 1")
	}

	if len(validationErrors) > 0 {
		logrus.Error("Configuration validation failed:")
		for _, err := range validationErrors {
			logrus.Errorf("   - %s", err)
		}
		return errors.NewAppErrorWithDetails(errors.ErrConfigValidation, "Configuration validation failed", strings.Join(validationErrors, "; "))
	}

	return nil
}

// DisplayConfig displays current configuration information
func (m *Manager) DisplayConfig() {
	logrus.Info("Current Configuration:")
	logrus.Infof("   Server: %s:%d", m.config.Server.Host, m.config.Server.Port)
	logrus.Infof("   API Keys loaded: %d", len(m.config.Keys.APIKeys))
	logrus.Infof("   Start index: %d", m.config.Keys.StartIndex)
	logrus.Infof("   Blacklist threshold: %d errors", m.config.Keys.BlacklistThreshold)
	logrus.Infof("   Max retries: %d", m.config.Keys.MaxRetries)
	logrus.Infof("   Upstream URLs: %s", strings.Join(m.config.OpenAI.BaseURLs, ", "))
	logrus.Infof("   Request timeout: %ds", m.config.OpenAI.RequestTimeout)
	logrus.Infof("   Response timeout: %ds", m.config.OpenAI.ResponseTimeout)
	logrus.Infof("   Idle connection timeout: %ds", m.config.OpenAI.IdleConnTimeout)

	authStatus := "disabled"
	if m.config.Auth.Enabled {
		authStatus = "enabled"
	}
	logrus.Infof("   Authentication: %s", authStatus)

	corsStatus := "disabled"
	if m.config.CORS.Enabled {
		corsStatus = "enabled"
	}
	logrus.Infof("   CORS: %s", corsStatus)
	logrus.Infof("   Max concurrent requests: %d", m.config.Performance.MaxConcurrentRequests)

	gzipStatus := "disabled"
	if m.config.Performance.EnableGzip {
		gzipStatus = "enabled"
	}
	logrus.Infof("   Gzip compression: %s", gzipStatus)

	requestLogStatus := "enabled"
	if !m.config.Log.EnableRequest {
		requestLogStatus = "disabled"
	}
	logrus.Infof("   Request logging: %s", requestLogStatus)
}

// Helper functions

// parseInteger parses integer environment variable
func parseInteger(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed
	}
	return defaultValue
}

// parseBoolean parses boolean environment variable
func parseBoolean(value string, defaultValue bool) bool {
	if value == "" {
		return defaultValue
	}

	lowerValue := strings.ToLower(value)
	switch lowerValue {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

// parseArray parses array environment variable (comma-separated)
func parseArray(value string, defaultValue []string) []string {
	if value == "" {
		return defaultValue
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}
	return result
}

// getEnvOrDefault gets environment variable or default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
