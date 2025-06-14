// Package keymanager provides high-performance API key management
package keymanager

import (
	"bufio"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gpt-load/internal/errors"
	"gpt-load/pkg/types"

	"github.com/sirupsen/logrus"
)

// Manager implements the KeyManager interface
type Manager struct {
	keysFilePath     string
	keys             []string
	keyPreviews      []string
	currentIndex     int64
	blacklistedKeys  sync.Map
	successCount     int64
	failureCount     int64
	keyFailureCounts sync.Map
	config           types.KeysConfig

	// Performance optimization: pre-compiled regex patterns
	permanentErrorPatterns []*regexp.Regexp

	// Memory management
	cleanupTicker *time.Ticker
	stopCleanup   chan bool

	// Read-write lock to protect key list
	keysMutex sync.RWMutex
}

// NewManager creates a new key manager
func NewManager(config types.KeysConfig) (types.KeyManager, error) {
	km := &Manager{
		currentIndex: int64(config.StartIndex),
		stopCleanup:  make(chan bool),
		config:       config,

		// Pre-compile regex patterns
		permanentErrorPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)invalid api key`),
			regexp.MustCompile(`(?i)incorrect api key`),
			regexp.MustCompile(`(?i)api key not found`),
			regexp.MustCompile(`(?i)unauthorized`),
			regexp.MustCompile(`(?i)account deactivated`),
			regexp.MustCompile(`(?i)billing`),
		},
	}

	// Start memory cleanup
	km.setupMemoryCleanup()

	// Load keys
	if len(config.APIKeys) == 0 {
		return nil, errors.NewAppError(errors.ErrNoKeysAvailable, "No API keys provided in environment variables")
	}

	if err := km.LoadKeys(); err != nil {
		return nil, err
	}

	return km, nil
}

// LoadKeys loads API keys from environment variables
func (km *Manager) LoadKeys() error {
	var keys []string
	var keyPreviews []string

	for _, key := range km.config.APIKeys {
		trimmedKey := strings.TrimSpace(key)
		if trimmedKey != "" {
			keys = append(keys, trimmedKey)
			// Create preview (first 8 chars + "..." + last 4 chars)
			if len(trimmedKey) > 12 {
				preview := trimmedKey[:8] + "..." + trimmedKey[len(trimmedKey)-4:]
				keyPreviews = append(keyPreviews, preview)
			} else {
				keyPreviews = append(keyPreviews, trimmedKey)
			}
		}
	}

	if len(keys) == 0 {
		return errors.NewAppError(errors.ErrNoKeysAvailable, "No valid API keys found in environment variables")
	}

	km.keysMutex.Lock()
	km.keys = keys
	km.keyPreviews = keyPreviews
	km.keysMutex.Unlock()

	logrus.Infof("Successfully loaded %d API keys from environment variables", len(keys))
	return nil
}

// GetNextKey gets the next available key (high-performance version)
func (km *Manager) GetNextKey() (*types.KeyInfo, error) {
	km.keysMutex.RLock()
	keysLen := len(km.keys)
	if keysLen == 0 {
		km.keysMutex.RUnlock()
		return nil, errors.ErrNoAPIKeysAvailable
	}

	// Fast path: directly get next key, avoid blacklist check overhead
	currentIdx := atomic.AddInt64(&km.currentIndex, 1) - 1
	keyIndex := int(currentIdx) % keysLen
	selectedKey := km.keys[keyIndex]
	keyPreview := km.keyPreviews[keyIndex]
	km.keysMutex.RUnlock()

	// Check if blacklisted
	if _, blacklisted := km.blacklistedKeys.Load(selectedKey); !blacklisted {
		return &types.KeyInfo{
			Key:     selectedKey,
			Index:   keyIndex,
			Preview: keyPreview,
		}, nil
	}

	// Slow path: find next available key
	return km.findNextAvailableKey(keyIndex, keysLen)
}

// findNextAvailableKey finds the next available non-blacklisted key
func (km *Manager) findNextAvailableKey(startIndex, keysLen int) (*types.KeyInfo, error) {
	km.keysMutex.RLock()
	defer km.keysMutex.RUnlock()

	blacklistedCount := 0
	for i := 0; i < keysLen; i++ {
		keyIndex := (startIndex + i) % keysLen
		selectedKey := km.keys[keyIndex]

		if _, blacklisted := km.blacklistedKeys.Load(selectedKey); !blacklisted {
			return &types.KeyInfo{
				Key:     selectedKey,
				Index:   keyIndex,
				Preview: km.keyPreviews[keyIndex],
			}, nil
		}
		blacklistedCount++
	}

	if blacklistedCount >= keysLen {
		logrus.Warn("All keys are blacklisted, resetting blacklist")
		km.blacklistedKeys = sync.Map{}
		km.keyFailureCounts = sync.Map{}

		// Return first key after reset
		firstKey := km.keys[0]
		firstPreview := km.keyPreviews[0]

		return &types.KeyInfo{
			Key:     firstKey,
			Index:   0,
			Preview: firstPreview,
		}, nil
	}

	return nil, errors.ErrAllAPIKeysBlacklisted
}

// RecordSuccess records successful key usage
func (km *Manager) RecordSuccess(key string) {
	atomic.AddInt64(&km.successCount, 1)
	// Reset failure count for this key on success
	km.keyFailureCounts.Delete(key)
}

// RecordFailure records key failure and potentially blacklists it
func (km *Manager) RecordFailure(key string, err error) {
	atomic.AddInt64(&km.failureCount, 1)

	// Check if this is a permanent error
	if km.isPermanentError(err) {
		km.blacklistedKeys.Store(key, time.Now())
		logrus.Debugf("Key blacklisted due to permanent error: %v", err)
		return
	}

	// Increment failure count
	failCount, _ := km.keyFailureCounts.LoadOrStore(key, new(int64))
	if counter, ok := failCount.(*int64); ok {
		newFailCount := atomic.AddInt64(counter, 1)

		// Blacklist if threshold exceeded
		if int(newFailCount) >= km.config.BlacklistThreshold {
			km.blacklistedKeys.Store(key, time.Now())
			logrus.Debugf("Key blacklisted after %d failures", newFailCount)
		}
	}
}

// isPermanentError checks if an error is permanent
func (km *Manager) isPermanentError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())
	for _, pattern := range km.permanentErrorPatterns {
		if pattern.MatchString(errorStr) {
			return true
		}
	}
	return false
}

// GetStats returns current statistics
func (km *Manager) GetStats() types.Stats {
	km.keysMutex.RLock()
	totalKeys := len(km.keys)
	km.keysMutex.RUnlock()

	blacklistedCount := 0
	km.blacklistedKeys.Range(func(key, value any) bool {
		blacklistedCount++
		return true
	})

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return types.Stats{
		CurrentIndex:    atomic.LoadInt64(&km.currentIndex),
		TotalKeys:       totalKeys,
		HealthyKeys:     totalKeys - blacklistedCount,
		BlacklistedKeys: blacklistedCount,
		SuccessCount:    atomic.LoadInt64(&km.successCount),
		FailureCount:    atomic.LoadInt64(&km.failureCount),
		MemoryUsage: types.MemoryUsage{
			Alloc:        m.Alloc,
			TotalAlloc:   m.TotalAlloc,
			Sys:          m.Sys,
			NumGC:        m.NumGC,
			LastGCTime:   time.Unix(0, int64(m.LastGC)).Format("2006-01-02 15:04:05"),
			NextGCTarget: m.NextGC,
		},
	}
}

// ResetBlacklist resets the blacklist
func (km *Manager) ResetBlacklist() {
	km.blacklistedKeys = sync.Map{}
	km.keyFailureCounts = sync.Map{}
	logrus.Info("Blacklist reset successfully")
}

// GetBlacklist returns current blacklisted keys
func (km *Manager) GetBlacklist() []types.BlacklistEntry {
	var blacklist []types.BlacklistEntry

	km.blacklistedKeys.Range(func(key, value any) bool {
		keyStr := key.(string)
		blacklistTime := value.(time.Time)

		// Create preview
		preview := keyStr
		if len(keyStr) > 12 {
			preview = keyStr[:8] + "..." + keyStr[len(keyStr)-4:]
		}

		// Get failure count
		failCount := 0
		if count, exists := km.keyFailureCounts.Load(keyStr); exists {
			failCount = int(atomic.LoadInt64(count.(*int64)))
		}

		blacklist = append(blacklist, types.BlacklistEntry{
			Key:         keyStr,
			Preview:     preview,
			Reason:      "Exceeded failure threshold",
			BlacklistAt: blacklistTime,
			FailCount:   failCount,
		})
		return true
	})

	return blacklist
}

// setupMemoryCleanup sets up periodic memory cleanup
func (km *Manager) setupMemoryCleanup() {
	// Reduce GC frequency to every 15 minutes to avoid performance impact
	km.cleanupTicker = time.NewTicker(15 * time.Minute)
	go func() {
		for {
			select {
			case <-km.cleanupTicker.C:
				// Only trigger GC if memory usage is high
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				// Trigger GC only if allocated memory is above 100MB
				if m.Alloc > 100*1024*1024 {
					runtime.GC()
					logrus.Debugf("Manual GC triggered, memory usage: %d MB", m.Alloc/1024/1024)
				}
			case <-km.stopCleanup:
				return
			}
		}
	}()
}

// Close closes the key manager and cleans up resources
func (km *Manager) Close() {
	if km.cleanupTicker != nil {
		km.cleanupTicker.Stop()
	}
	close(km.stopCleanup)
}
