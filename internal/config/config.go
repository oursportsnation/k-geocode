package config

import (
	"fmt"
	"os"
	"strings"
	"time"
	
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Providers ProvidersConfig `yaml:"providers"`
	Redis     RedisConfig     `yaml:"redis"`
	Logging   LoggingConfig   `yaml:"logging"`
	API       APIConfig       `yaml:"api"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port               string        `yaml:"port"`
	ReadTimeout        time.Duration `yaml:"read_timeout"`
	WriteTimeout       time.Duration `yaml:"write_timeout"`
	MaxRequestBodySize string        `yaml:"max_request_body_size"`
}

// ProvidersConfig represents providers configuration
type ProvidersConfig struct {
	VWorld ProviderConfig `yaml:"vworld"`
	Kakao  ProviderConfig `yaml:"kakao"`
}

// ProviderConfig represents individual provider configuration
type ProviderConfig struct {
	Enabled        bool                  `yaml:"enabled"`
	APIKey         string                `yaml:"api_key"`
	DailyLimit     int                   `yaml:"daily_limit"`
	Timeout        time.Duration         `yaml:"timeout"`
	CircuitBreaker CircuitBreakerConfig  `yaml:"circuit_breaker"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           `yaml:"failure_threshold"`
	SuccessThreshold int           `yaml:"success_threshold"`
	Timeout          time.Duration `yaml:"timeout"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Addr     string        `yaml:"addr"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	Timeout  time.Duration `yaml:"timeout"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// APIConfig represents API configuration
type APIConfig struct {
	MaxBatchSize    int           `yaml:"max_batch_size"`
	RequestTimeout  time.Duration `yaml:"request_timeout"`
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	// 파일 읽기
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// 환경변수 치환
	data = []byte(expandEnv(string(data)))
	
	// YAML 파싱
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// 기본값 설정
	setDefaults(&config)
	
	// 검증
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	return &config, nil
}

// expandEnv replaces ${VAR} or $VAR with environment variables
func expandEnv(s string) string {
	return os.Expand(s, func(key string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		// 환경변수가 없으면 기본값 사용
		switch key {
		case "SERVER_PORT":
			return "8080"
		case "REDIS_ADDR":
			return "localhost:6379"
		case "LOG_LEVEL":
			return "info"
		default:
			return ""
		}
	})
}

// setDefaults sets default values for configuration
func setDefaults(cfg *Config) {
	// Server defaults
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 15 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 15 * time.Second
	}
	if cfg.Server.MaxRequestBodySize == "" {
		cfg.Server.MaxRequestBodySize = "1MB"
	}
	
	// Provider defaults
	if cfg.Providers.VWorld.Timeout == 0 {
		cfg.Providers.VWorld.Timeout = 5 * time.Second
	}
	if cfg.Providers.Kakao.Timeout == 0 {
		cfg.Providers.Kakao.Timeout = 5 * time.Second
	}
	
	// Circuit Breaker defaults
	if cfg.Providers.VWorld.CircuitBreaker.FailureThreshold == 0 {
		cfg.Providers.VWorld.CircuitBreaker.FailureThreshold = 5
	}
	if cfg.Providers.VWorld.CircuitBreaker.SuccessThreshold == 0 {
		cfg.Providers.VWorld.CircuitBreaker.SuccessThreshold = 2
	}
	if cfg.Providers.VWorld.CircuitBreaker.Timeout == 0 {
		cfg.Providers.VWorld.CircuitBreaker.Timeout = 60 * time.Second
	}
	
	// Same for Kakao
	if cfg.Providers.Kakao.CircuitBreaker.FailureThreshold == 0 {
		cfg.Providers.Kakao.CircuitBreaker.FailureThreshold = 5
	}
	if cfg.Providers.Kakao.CircuitBreaker.SuccessThreshold == 0 {
		cfg.Providers.Kakao.CircuitBreaker.SuccessThreshold = 2
	}
	if cfg.Providers.Kakao.CircuitBreaker.Timeout == 0 {
		cfg.Providers.Kakao.CircuitBreaker.Timeout = 60 * time.Second
	}
	
	// Redis defaults
	if cfg.Redis.Timeout == 0 {
		cfg.Redis.Timeout = 5 * time.Second
	}
	
	// Logging defaults
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stdout"
	}
	
	// API defaults
	if cfg.API.MaxBatchSize == 0 {
		cfg.API.MaxBatchSize = 100
	}
	if cfg.API.RequestTimeout == 0 {
		cfg.API.RequestTimeout = 15 * time.Second
	}
}

// validate validates configuration
func validate(cfg *Config) error {
	// Port 검증
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	
	// Provider 검증
	if cfg.Providers.VWorld.Enabled && cfg.Providers.VWorld.APIKey == "" {
		return fmt.Errorf("vWorld API key is required when enabled")
	}
	if cfg.Providers.Kakao.Enabled && cfg.Providers.Kakao.APIKey == "" {
		return fmt.Errorf("Kakao API key is required when enabled")
	}
	
	// 최소 하나의 Provider는 활성화되어야 함
	if !cfg.Providers.VWorld.Enabled && !cfg.Providers.Kakao.Enabled {
		return fmt.Errorf("at least one provider must be enabled")
	}
	
	// Redis 검증
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("redis address is required")
	}
	
	// API 검증
	if cfg.API.MaxBatchSize < 1 || cfg.API.MaxBatchSize > 1000 {
		return fmt.Errorf("max_batch_size must be between 1 and 1000")
	}
	
	return nil
}

// LoadWithEnv loads configuration with environment-specific overrides
func LoadWithEnv(basePath string, env string) (*Config, error) {
	// 기본 설정 로드
	config, err := Load(basePath)
	if err != nil {
		return nil, err
	}
	
	// 환경별 설정 파일이 있으면 오버라이드
	if env != "" {
		envPath := strings.Replace(basePath, ".yaml", "."+env+".yaml", 1)
		if _, err := os.Stat(envPath); err == nil {
			// 환경별 파일 읽기
			data, err := os.ReadFile(envPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read env config file: %w", err)
			}
			
			// 환경변수 치환
			data = []byte(expandEnv(string(data)))
			
			// YAML 파싱
			var envConfig Config
			if err := yaml.Unmarshal(data, &envConfig); err != nil {
				return nil, fmt.Errorf("failed to parse env config file: %w", err)
			}
			
			// 환경별 설정으로 오버라이드
			mergeConfig(config, &envConfig)
			
			// 기본값 재설정 및 검증
			setDefaults(config)
			if err := validate(config); err != nil {
				return nil, fmt.Errorf("failed to load env config: %w", err)
			}
		}
	}
	
	return config, nil
}

// mergeConfig merges environment-specific config into base config
func mergeConfig(base, override *Config) {
	// 간단한 구현 - 실제로는 더 복잡한 deep merge가 필요할 수 있음
	if override.Server.Port != "" {
		base.Server.Port = override.Server.Port
	}
	if override.Logging.Level != "" {
		base.Logging.Level = override.Logging.Level
	}
	if override.Logging.Format != "" {
		base.Logging.Format = override.Logging.Format
	}
	// 필요한 다른 필드들도 추가
}