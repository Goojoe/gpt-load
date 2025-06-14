// Package types defines common interfaces and types used across the application
package types

import (
	"time"

	"github.com/gin-gonic/gin"
)

// ConfigManager defines the interface for configuration management
type ConfigManager interface {
	GetServerConfig() ServerConfig
	GetKeysConfig() KeysConfig
	GetOpenAIConfig() OpenAIConfig
	GetAuthConfig() AuthConfig
	GetCORSConfig() CORSConfig
	GetPerformanceConfig() PerformanceConfig
	GetLogConfig() LogConfig
	Validate() error
	DisplayConfig()
}

// KeyManager defines the interface for API key management
type KeyManager interface {
	LoadKeys() error
	GetNextKey() (*KeyInfo, error)
	RecordSuccess(key string)
	RecordFailure(key string, err error)
	GetStats() Stats
	ResetBlacklist()
	GetBlacklist() []BlacklistEntry
	Close()
}

// ProxyServer defines the interface for proxy server
type ProxyServer interface {
	HandleProxy(c *gin.Context)
	Close()
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port                    int    `json:"port"`
	Host                    string `json:"host"`
	ReadTimeout             int    `json:"readTimeout"`
	WriteTimeout            int    `json:"writeTimeout"`
	IdleTimeout             int    `json:"idleTimeout"`
	GracefulShutdownTimeout int    `json:"gracefulShutdownTimeout"`
}

// KeysConfig represents keys configuration
type KeysConfig struct {
	APIKeys            []string `json:"apiKeys"`
	StartIndex         int      `json:"startIndex"`
	BlacklistThreshold int      `json:"blacklistThreshold"`
	MaxRetries         int      `json:"maxRetries"`
}

// OpenAIConfig represents OpenAI API configuration
type OpenAIConfig struct {
	BaseURL         string   `json:"baseUrl"`
	BaseURLs        []string `json:"baseUrls"`
	RequestTimeout  int      `json:"requestTimeout"`
	ResponseTimeout int      `json:"responseTimeout"`
	IdleConnTimeout int      `json:"idleConnTimeout"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled          bool     `json:"enabled"`
	AllowedOrigins   []string `json:"allowedOrigins"`
	AllowedMethods   []string `json:"allowedMethods"`
	AllowedHeaders   []string `json:"allowedHeaders"`
	AllowCredentials bool     `json:"allowCredentials"`
}

// PerformanceConfig represents performance configuration
type PerformanceConfig struct {
	MaxConcurrentRequests int  `json:"maxConcurrentRequests"`
	EnableGzip            bool `json:"enableGzip"`
}

// LogConfig represents logging configuration
type LogConfig struct {
	Level         string `json:"level"`
	Format        string `json:"format"`
	EnableFile    bool   `json:"enableFile"`
	FilePath      string `json:"filePath"`
	EnableRequest bool   `json:"enableRequest"`
}

// KeyInfo represents API key information
type KeyInfo struct {
	Key     string `json:"key"`
	Index   int    `json:"index"`
	Preview string `json:"preview"`
}

// Stats represents system statistics
type Stats struct {
	CurrentIndex    int64       `json:"currentIndex"`
	TotalKeys       int         `json:"totalKeys"`
	HealthyKeys     int         `json:"healthyKeys"`
	BlacklistedKeys int         `json:"blacklistedKeys"`
	SuccessCount    int64       `json:"successCount"`
	FailureCount    int64       `json:"failureCount"`
	MemoryUsage     MemoryUsage `json:"memoryUsage"`
}

// MemoryUsage represents memory usage statistics
type MemoryUsage struct {
	Alloc        uint64 `json:"alloc"`
	TotalAlloc   uint64 `json:"totalAlloc"`
	Sys          uint64 `json:"sys"`
	NumGC        uint32 `json:"numGC"`
	LastGCTime   string `json:"lastGCTime"`
	NextGCTarget uint64 `json:"nextGCTarget"`
}

// BlacklistEntry represents a blacklisted key entry
type BlacklistEntry struct {
	Key         string    `json:"key"`
	Preview     string    `json:"preview"`
	Reason      string    `json:"reason"`
	BlacklistAt time.Time `json:"blacklistAt"`
	FailCount   int       `json:"failCount"`
}

// RetryError represents retry error information
type RetryError struct {
	StatusCode   int    `json:"statusCode"`
	ErrorMessage string `json:"errorMessage"`
	KeyIndex     int    `json:"keyIndex"`
	Attempt      int    `json:"attempt"`
}
