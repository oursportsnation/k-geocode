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

// Config k-geocode 클라이언트 설정
type Config struct {
	// VWorldAPIKey vWorld API 키
	// https://www.vworld.kr 에서 발급
	VWorldAPIKey string

	// KakaoAPIKey Kakao REST API 키
	// https://developers.kakao.com 에서 발급
	KakaoAPIKey string

	// Timeout HTTP 요청 타임아웃 (기본: 5초)
	Timeout time.Duration

	// MaxRetries 재시도 횟수 (기본: 2)
	// 현재는 사용되지 않지만 향후 확장 예정
	MaxRetries int

	// LogLevel 로그 레벨 (기본: "info")
	// 옵션: "debug", "info", "warn", "error"
	LogLevel string

	// ConcurrentLimit 배치 처리 시 동시 실행 제한 (기본: 10)
	ConcurrentLimit int
}

// DefaultConfig 기본 설정을 반환합니다
func DefaultConfig() Config {
	return Config{
		Timeout:         5 * time.Second,
		MaxRetries:      2,
		LogLevel:        "info",
		ConcurrentLimit: 10,
	}
}

// Validate 설정값을 검증합니다
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

// SetDefaults 기본값을 설정합니다 (0값인 필드만)
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
