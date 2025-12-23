// Copyright 2025 Our Sports Nation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package geocoding

import (
	"fmt"
	"time"
)

// Config holds the configuration for the geocoding client.
type Config struct {
	// VWorldAPIKey is the API key for vWorld geocoding service.
	// Obtain from https://www.vworld.kr
	VWorldAPIKey string

	// KakaoAPIKey is the REST API key for Kakao geocoding service.
	// Obtain from https://developers.kakao.com
	KakaoAPIKey string

	// Timeout is the HTTP request timeout. Default: 5 seconds.
	Timeout time.Duration

	// MaxRetries is the number of retry attempts. Default: 2.
	// Reserved for future use.
	MaxRetries int

	// LogLevel sets the logging verbosity. Default: "info".
	// Valid values: "debug", "info", "warn", "error".
	LogLevel string

	// ConcurrentLimit is the maximum concurrent requests for batch operations. Default: 10.
	ConcurrentLimit int
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() Config {
	return Config{
		Timeout:         5 * time.Second,
		MaxRetries:      2,
		LogLevel:        "info",
		ConcurrentLimit: 10,
	}
}

// Validate checks that the configuration is valid.
// It returns an error if required fields are missing or values are out of range.
func (c *Config) Validate() error {
	// 최소 하나의 API 키는 필수
	if c.VWorldAPIKey == "" && c.KakaoAPIKey == "" {
		return fmt.Errorf("at least one API key (VWorldAPIKey or KakaoAPIKey) is required")
	}

	// Timeout 검증
	if c.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	// MaxRetries 검증
	if c.MaxRetries < 0 {
		return fmt.Errorf("maxRetries cannot be negative")
	}

	// ConcurrentLimit 검증
	if c.ConcurrentLimit < 1 {
		return fmt.Errorf("concurrentLimit must be at least 1")
	}

	if c.ConcurrentLimit > 100 {
		return fmt.Errorf("concurrentLimit cannot exceed 100")
	}

	// LogLevel 검증
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if c.LogLevel != "" && !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", c.LogLevel)
	}

	return nil
}

// SetDefaults applies default values to zero-valued fields.
func (c *Config) SetDefaults() {
	if c.Timeout == 0 {
		c.Timeout = 5 * time.Second
	}

	if c.MaxRetries == 0 {
		c.MaxRetries = 2
	}

	if c.LogLevel == "" {
		c.LogLevel = "info"
	}

	if c.ConcurrentLimit == 0 {
		c.ConcurrentLimit = 10
	}
}
