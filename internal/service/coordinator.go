package service

import (
	"context"
	"fmt"
	"github.com/oursportsnation/k-geocode/internal/config"
	"github.com/oursportsnation/k-geocode/internal/provider"
	"github.com/oursportsnation/k-geocode/pkg/httpclient"

	"go.uber.org/zap"
)

// CoordinatorInterface 코디네이터 인터페이스
type CoordinatorInterface interface {
	HealthCheck(ctx context.Context) HealthStatus
	GetGeocodingService() *GeocodingService
}

// Coordinator 서비스 조율자 - 모든 서비스와 Provider를 초기화하고 관리
type Coordinator struct {
	config           *config.Config
	geocodingService *GeocodingService
	providers        []provider.GeocodingProvider
	logger           *zap.Logger
}

// NewCoordinator 조율자 생성자
func NewCoordinator(cfg *config.Config, logger *zap.Logger) (*Coordinator, error) {
	coord := &Coordinator{
		config: cfg,
		logger: logger,
	}
	
	// Provider 초기화
	if err := coord.initProviders(); err != nil {
		return nil, fmt.Errorf("failed to initialize providers: %w", err)
	}
	
	// 서비스 초기화
	coord.initServices()
	
	return coord, nil
}

// initProviders Provider들을 초기화
func (c *Coordinator) initProviders() error {
	c.providers = make([]provider.GeocodingProvider, 0)
	
	// HTTP 클라이언트 생성
	httpClient := httpclient.DefaultClient()
	
	// vWorld Provider
	if c.config.Providers.VWorld.Enabled {
		if c.config.Providers.VWorld.APIKey == "" {
			c.logger.Warn("vWorld provider is enabled but API key is missing")
		} else {
			vworldProvider := provider.NewVWorldProvider(
				c.config.Providers.VWorld.APIKey,
				httpClient,
				c.logger.Named("vworld"),
			)
			c.providers = append(c.providers, vworldProvider)
			c.logger.Info("vWorld provider initialized")
		}
	}
	
	// Kakao Provider
	if c.config.Providers.Kakao.Enabled {
		if c.config.Providers.Kakao.APIKey == "" {
			c.logger.Warn("Kakao provider is enabled but API key is missing")
		} else {
			kakaoProvider := provider.NewKakaoProvider(
				c.config.Providers.Kakao.APIKey,
				httpClient,
				c.logger.Named("kakao"),
			)
			c.providers = append(c.providers, kakaoProvider)
			c.logger.Info("Kakao provider initialized")
		}
	}
	
	// 최소 하나의 Provider는 필요
	if len(c.providers) == 0 {
		return fmt.Errorf("no providers available - check API keys")
	}
	
	c.logger.Info("Providers initialized",
		zap.Int("count", len(c.providers)),
	)
	
	return nil
}

// initServices 서비스들을 초기화
func (c *Coordinator) initServices() {
	// 지오코딩 서비스 초기화
	c.geocodingService = NewGeocodingService(c.providers, c.logger.Named("geocoding"))
	
	c.logger.Info("Services initialized")
}

// GetGeocodingService 지오코딩 서비스 반환
func (c *Coordinator) GetGeocodingService() *GeocodingService {
	return c.geocodingService
}

// GetProviders Provider 목록 반환
func (c *Coordinator) GetProviders() []provider.GeocodingProvider {
	return c.providers
}

// HealthCheck 시스템 헬스 체크
func (c *Coordinator) HealthCheck(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Healthy:   true,
		Providers: make([]ProviderStatus, 0),
	}
	
	// 각 Provider의 가용성 확인
	for _, p := range c.providers {
		providerStatus := ProviderStatus{
			Name:      p.Name(),
			Available: p.IsAvailable(ctx),
		}
		
		status.Providers = append(status.Providers, providerStatus)
		
		// 하나라도 사용 가능하면 시스템은 healthy
		if providerStatus.Available {
			status.Healthy = true
		}
	}
	
	// 모든 Provider가 사용 불가능하면 unhealthy
	allUnavailable := true
	for _, ps := range status.Providers {
		if ps.Available {
			allUnavailable = false
			break
		}
	}
	if allUnavailable {
		status.Healthy = false
	}
	
	return status
}

// Shutdown 조율자 종료
func (c *Coordinator) Shutdown() error {
	c.logger.Info("Shutting down coordinator")
	
	// 필요한 정리 작업 수행
	// 예: Provider 연결 종료, 리소스 해제 등
	
	return nil
}

// HealthStatus 헬스 체크 상태
type HealthStatus struct {
	Healthy   bool             `json:"healthy"`
	Providers []ProviderStatus `json:"providers"`
}

// ProviderStatus Provider 상태
type ProviderStatus struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}