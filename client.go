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
	"context"
	"fmt"
	"strings"

	"github.com/oursportsnation/k-geocode/internal/provider"
	"github.com/oursportsnation/k-geocode/internal/service"
	"github.com/oursportsnation/k-geocode/pkg/httpclient"
	"github.com/oursportsnation/k-geocode/pkg/logger"
)

// Client is the k-geocode geocoding client that provides unified access
// to multiple Korean geocoding providers with automatic fallback.
type Client struct {
	service   *service.GeocodingService
	providers []provider.GeocodingProvider
	config    Config
}

// New creates a new geocoding client with the given configuration.
// At least one API key (VWorldAPIKey or KakaoAPIKey) must be provided.
func New(cfg Config) (*Client, error) {
	// 설정 검증
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// 기본값 적용
	cfg.SetDefaults()

	// Logger 초기화
	log, err := logger.New(cfg.LogLevel, "json")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// HTTP 클라이언트 생성
	httpClient := httpclient.NewClient(cfg.Timeout)

	// Provider들 초기화
	var providers []provider.GeocodingProvider

	// vWorld Provider(s) - 콤마로 구분된 여러 키 지원
	if cfg.VWorldAPIKey != "" {
		vworldKeys := strings.Split(cfg.VWorldAPIKey, ",")
		for i, key := range vworldKeys {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			vworldProvider := provider.NewVWorldProvider(key, httpClient, log)
			providers = append(providers, vworldProvider)
			log.Info(fmt.Sprintf("vWorld provider #%d registered", i+1))
		}
	}

	// Kakao Provider
	if cfg.KakaoAPIKey != "" {
		kakaoProvider := provider.NewKakaoProvider(cfg.KakaoAPIKey, httpClient, log)
		providers = append(providers, kakaoProvider)
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one API key (VWorld or Kakao) is required")
	}

	// 지오코딩 서비스 생성
	geocodingService := service.NewGeocodingService(providers, log)

	return &Client{
		service:   geocodingService,
		providers: providers,
		config:    cfg,
	}, nil
}

// Geocode converts a Korean address to WGS84 coordinates.
// It automatically falls back through providers (vWorld → Kakao) and
// address types (ROAD → PARCEL) until a result is found.
func (c *Client) Geocode(ctx context.Context, address string) (*Result, error) {
	return c.GeocodeWithType(ctx, address, "")
}

// GeocodeWithType converts a Korean address to WGS84 coordinates
// using the specified address type.
//
// Use [AddressTypeRoad] for road-based addresses (도로명) or
// [AddressTypeParcel] for parcel-based addresses (지번).
// Pass an empty string to automatically try ROAD then PARCEL.
func (c *Client) GeocodeWithType(ctx context.Context, address string, addressType AddressType) (*Result, error) {
	resp, err := c.service.Geocode(ctx, address, string(addressType))
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("geocoding failed: %s", resp.Error)
	}

	// 내부 응답을 공개 타입으로 변환
	result := &Result{
		Latitude:  resp.Coordinate.Latitude,
		Longitude: resp.Coordinate.Longitude,
		Provider:  resp.Provider,
	}

	// 주소 상세 정보가 있으면 추가
	if resp.AddressDetail != nil {
		result.AddressDetail = &AddressDetail{
			RoadAddress:   resp.AddressDetail.RoadAddress,
			ParcelAddress: resp.AddressDetail.ParcelAddress,
			BuildingName:  resp.AddressDetail.BuildingName,
			Zipcode:       resp.AddressDetail.Zipcode,
		}
	}

	// Provider 시도 내역
	for _, attempt := range resp.Attempts {
		result.Attempts = append(result.Attempts, Attempt{
			Provider: attempt.Provider,
			Success:  attempt.Success,
			Error:    attempt.Error,
		})
	}

	return result, nil
}

// GeocodeBatch converts multiple addresses concurrently (max 100).
// Up to 10 addresses are processed in parallel.
// Partial failures are allowed; successful results are returned alongside nil entries for failures.
func (c *Client) GeocodeBatch(ctx context.Context, addresses []string) ([]*Result, error) {
	if len(addresses) == 0 {
		return []*Result{}, nil
	}

	if len(addresses) > 100 {
		return nil, fmt.Errorf("too many addresses: maximum 100, got %d", len(addresses))
	}

	bulkResp, err := c.service.GeocodeBatch(ctx, addresses)
	if err != nil {
		return nil, err
	}

	// 내부 응답을 공개 타입으로 변환
	results := make([]*Result, 0, len(bulkResp.Results))
	for _, resp := range bulkResp.Results {
		if !resp.Success {
			// 실패한 경우 nil 추가
			results = append(results, nil)
			continue
		}

		result := &Result{
			Latitude:  resp.Coordinate.Latitude,
			Longitude: resp.Coordinate.Longitude,
			Provider:  resp.Provider,
		}

		if resp.AddressDetail != nil {
			result.AddressDetail = &AddressDetail{
				RoadAddress:   resp.AddressDetail.RoadAddress,
				ParcelAddress: resp.AddressDetail.ParcelAddress,
				BuildingName:  resp.AddressDetail.BuildingName,
				Zipcode:       resp.AddressDetail.Zipcode,
			}
		}

		results = append(results, result)
	}

	return results, nil
}

// Close releases any resources held by the client.
func (c *Client) Close() error {
	// 현재는 정리할 리소스 없음
	// 향후 Connection Pool 등이 추가되면 여기서 정리
	return nil
}

// IsAvailable returns true if at least one geocoding provider is available.
func (c *Client) IsAvailable(ctx context.Context) bool {
	for _, p := range c.providers {
		if p.IsAvailable(ctx) {
			return true
		}
	}
	return false
}

// GetProviders returns the list of configured provider names.
func (c *Client) GetProviders() []string {
	names := make([]string, 0, len(c.providers))
	for _, p := range c.providers {
		names = append(names, p.Name())
	}
	return names
}
