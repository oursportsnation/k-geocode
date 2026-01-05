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

package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/oursportsnation/k-geocode/internal/provider"
	"github.com/oursportsnation/k-geocode/internal/utils"

	"go.uber.org/zap"
)

// GeocodingServiceInterface 지오코딩 서비스 인터페이스
type GeocodingServiceInterface interface {
	Geocode(ctx context.Context, address string, addressType string) (*model.GeocodingResponse, error)
	GeocodeBatch(ctx context.Context, addresses []string) (*model.BulkResponse, error)
}

// GeocodingService 지오코딩 서비스
type GeocodingService struct {
	providers []provider.GeocodingProvider
	logger    *zap.Logger
}

// NewGeocodingService 지오코딩 서비스 생성자
func NewGeocodingService(providers []provider.GeocodingProvider, logger *zap.Logger) *GeocodingService {
	return &GeocodingService{
		providers: providers,
		logger:    logger,
	}
}

// Geocode 주소를 좌표로 변환 (단건)
func (s *GeocodingService) Geocode(ctx context.Context, address string, addressType string) (*model.GeocodingResponse, error) {
	start := time.Now()

	// 1. 입력 검증
	address = utils.NormalizeAddress(address)
	if !utils.IsValidAddress(address) {
		s.logger.Warn("Invalid address format",
			zap.String("address", address),
		)
		return &model.GeocodingResponse{
			Success:        false,
			Error:          "invalid address format",
			ProcessedAt:    time.Now(),
			ProcessingTime: time.Since(start),
		}, nil
	}

	s.logger.Info("Starting geocoding",
		zap.String("address", address),
		zap.String("address_type", addressType),
		zap.Int("providers", len(s.providers)),
	)

	// Provider 시도 내역 추적
	var attempts []model.ProviderAttempt

	// 2. Provider 순회 (폴백)
	for i, p := range s.providers {
		if !p.IsAvailable(ctx) {
			s.logger.Debug("Provider not available",
				zap.String("provider", p.Name()),
			)
			// 사용 불가능한 Provider도 기록
			attempts = append(attempts, model.ProviderAttempt{
				Provider: p.Name(),
				Success:  false,
				Error:    "provider not available",
			})
			continue
		}

		s.logger.Debug("Trying provider",
			zap.String("provider", p.Name()),
			zap.Int("attempt", i+1),
		)

		// Provider 호출
		var result *model.ProviderResult
		var err error

		// vWorld Provider이고 주소 타입이 지정된 경우
		if vworldProvider, ok := p.(*provider.VWorldProvider); ok && addressType != "" {
			result, err = vworldProvider.GeocodeWithType(ctx, address, addressType)
		} else {
			result, err = p.Geocode(ctx, address)
		}

		// 시스템 에러 처리
		if err != nil {
			// 분류된 에러인 경우
			if ce, ok := provider.IsClassifiedError(err); ok {
				s.logger.Warn("Provider error",
					zap.String("provider", p.Name()),
					zap.String("error_type", ce.Type.String()),
					zap.Error(err),
				)

				// 시도 내역 기록
				attempts = append(attempts, model.ProviderAttempt{
					Provider: p.Name(),
					Success:  false,
					Error:    err.Error(),
				})

				// 인증 실패 또는 한도 초과 시 Provider 비활성화 후 폴백
				if ce.Type == provider.ErrorTypeUnauthorized {
					p.Disable(fmt.Sprintf("Authentication failed: %s", err.Error()))
					s.logger.Error("Provider disabled due to authentication failure",
						zap.String("provider", p.Name()),
						zap.String("reason", err.Error()),
					)
					continue
				}
				if ce.Type == provider.ErrorTypeRateLimitExceeded {
					p.Disable(fmt.Sprintf("Rate limit exceeded: %s", err.Error()))
					s.logger.Warn("Provider disabled due to rate limit",
						zap.String("provider", p.Name()),
						zap.String("reason", err.Error()),
					)
					continue
				}

				// 기타 폴백 불가능한 에러는 즉시 반환
				if !ce.Fallback {
					return &model.GeocodingResponse{
						Success:        false,
						Provider:       p.Name(),
						Attempts:       attempts,
						Error:          err.Error(),
						ProcessedAt:    time.Now(),
						ProcessingTime: time.Since(start),
					}, nil
				}

				// 폴백 가능한 에러는 다음 Provider로
				continue
			}

			// 기타 에러
			s.logger.Error("Provider unexpected error",
				zap.String("provider", p.Name()),
				zap.Error(err),
			)

			// 시도 내역 기록
			attempts = append(attempts, model.ProviderAttempt{
				Provider: p.Name(),
				Success:  false,
				Error:    err.Error(),
			})
			continue
		}

		// 결과가 있는 경우
		if result != nil && result.Success {
			// 성공 시도 기록
			attempts = append(attempts, model.ProviderAttempt{
				Provider: p.Name(),
				Success:  true,
			})

			// 3. 좌표 정규화
			normalized := s.normalizeResponse(result, p.Name())
			normalized.ProcessedAt = time.Now()
			normalized.ProcessingTime = time.Since(start)
			normalized.Attempts = attempts

			s.logger.Info("Geocoding succeeded",
				zap.String("provider", p.Name()),
				zap.Float64("latitude", normalized.Coordinate.Latitude),
				zap.Float64("longitude", normalized.Coordinate.Longitude),
				zap.Duration("processing_time", normalized.ProcessingTime),
			)

			return normalized, nil
		}

		// 결과 없음 - 다음 Provider로
		s.logger.Debug("Provider returned no results",
			zap.String("provider", p.Name()),
		)

		// 시도 내역 기록
		attempts = append(attempts, model.ProviderAttempt{
			Provider: p.Name(),
			Success:  false,
			Error:    "address not found",
		})
	}
	
	// 4. 모든 Provider 실패
	s.logger.Warn("All providers failed to geocode",
		zap.String("address", address),
		zap.Duration("total_time", time.Since(start)),
	)

	return &model.GeocodingResponse{
		Success:        false,
		Provider:       "none",
		Attempts:       attempts,
		Error:          "all providers failed to geocode the address",
		ProcessedAt:    time.Now(),
		ProcessingTime: time.Since(start),
	}, nil
}

// GeocodeBatch 대량 주소 변환
func (s *GeocodingService) GeocodeBatch(ctx context.Context, addresses []string) (*model.BulkResponse, error) {
	start := time.Now()
	
	if len(addresses) == 0 {
		return &model.BulkResponse{
			Results:        []*model.GeocodingResponse{},
			ProcessingTime: 0,
		}, nil
	}
	
	s.logger.Info("Starting batch geocoding",
		zap.Int("addresses", len(addresses)),
	)
	
	// 결과 슬라이스 초기화
	results := make([]*model.GeocodingResponse, len(addresses))
	
	// 동시 처리를 위한 설정
	const maxConcurrent = 10 // 최대 동시 처리 수
	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	
	// 각 주소 처리
	for i, addr := range addresses {
		wg.Add(1)
		go func(idx int, address string) {
			defer wg.Done()
			
			// 동시 실행 제한
			sem <- struct{}{}
			defer func() { <-sem }()
			
			// 개별 지오코딩 (배치에서는 타입 지정 불가)
			result, err := s.Geocode(ctx, address, "")
			if err != nil {
				// 에러 발생 시에도 실패 결과를 기록
				results[idx] = &model.GeocodingResponse{
					Success:     false,
					Error:       err.Error(),
					ProcessedAt: time.Now(),
				}
			} else {
				results[idx] = result
			}
		}(i, addr)
	}
	
	// 모든 처리 완료 대기
	wg.Wait()
	
	// 통계 계산
	response := &model.BulkResponse{
		Results:        results,
		ProcessingTime: time.Since(start),
	}
	
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}
	
	response.Summary.Total = len(addresses)
	response.Summary.Success = successCount
	response.Summary.Failed = len(addresses) - successCount
	
	s.logger.Info("Batch geocoding completed",
		zap.Int("total", response.Summary.Total),
		zap.Int("success", response.Summary.Success),
		zap.Int("failed", response.Summary.Failed),
		zap.Duration("processing_time", response.ProcessingTime),
	)
	
	return response, nil
}

// normalizeResponse Provider 결과를 정규화된 응답으로 변환
func (s *GeocodingService) normalizeResponse(result *model.ProviderResult, providerName string) *model.GeocodingResponse {
	// 좌표 정규화 (소수점 6자리)
	normalizedCoord := model.Coordinate{
		Latitude:  utils.RoundToSixDecimal(result.Coordinate.Latitude),
		Longitude: utils.RoundToSixDecimal(result.Coordinate.Longitude),
	}
	
	// 좌표 유효성 검증
	if !utils.ValidateCoordinate(normalizedCoord.Latitude, normalizedCoord.Longitude) {
		s.logger.Warn("Invalid coordinates",
			zap.Float64("latitude", normalizedCoord.Latitude),
			zap.Float64("longitude", normalizedCoord.Longitude),
		)
		return &model.GeocodingResponse{
			Success:  false,
			Provider: providerName,
			Error:    "invalid coordinates",
		}
	}
	
	// 한국 영역 확인 (선택적)
	if !utils.IsValidKoreanCoordinate(normalizedCoord.Latitude, normalizedCoord.Longitude) {
		s.logger.Warn("Coordinates outside Korea",
			zap.Float64("latitude", normalizedCoord.Latitude),
			zap.Float64("longitude", normalizedCoord.Longitude),
		)
		// 경고만 하고 계속 진행
	}
	
	return &model.GeocodingResponse{
		Success:       true,
		Coordinate:    &normalizedCoord,
		AddressDetail: &result.AddressDetail,
		Provider:      providerName,
	}
}

// ValidateAddress 주소 유효성 검증 (외부 노출용)
func (s *GeocodingService) ValidateAddress(address string) error {
	normalized := utils.NormalizeAddress(address)
	if !utils.IsValidAddress(normalized) {
		return errors.New("invalid address format")
	}
	return nil
}

// GetAvailableProviders 사용 가능한 Provider 목록 반환
func (s *GeocodingService) GetAvailableProviders(ctx context.Context) []string {
	var available []string
	for _, p := range s.providers {
		if p.IsAvailable(ctx) {
			available = append(available, p.Name())
		}
	}
	return available
}