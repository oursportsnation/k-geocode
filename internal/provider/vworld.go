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

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/oursportsnation/k-geocode/pkg/httpclient"

	"go.uber.org/zap"
)

// VWorldProvider vWorld API 클라이언트
type VWorldProvider struct {
	apiKey        string
	httpClient    *httpclient.Client
	baseURL       string
	logger        *zap.Logger
	disabled      bool
	disableReason string
	mu            sync.RWMutex
}

// VWorldResponse vWorld API 응답 구조체
type VWorldResponse struct {
	Response struct {
		Status string `json:"status"`
		Result struct {
			CRS string `json:"crs"`
			Point struct {
				X string `json:"x"` // 경도
				Y string `json:"y"` // 위도
			} `json:"point"`
		} `json:"result"`
		Input struct {
			Type    string `json:"type"`
			Address string `json:"address"`
		} `json:"input"`
		Refined struct {
			Text string `json:"text"`
			Structure struct {
				Level0   string `json:"level0"`
				Level1   string `json:"level1"`
				Level2   string `json:"level2"`
				Level3   string `json:"level3"`
				Level4L  string `json:"level4L"`
				Level4LC string `json:"level4LC"`
				Level4A  string `json:"level4A"`
				Level4AC string `json:"level4AC"`
				Level5   string `json:"level5"`
				Detail   string `json:"detail"`
			} `json:"structure"`
		} `json:"refined"`
		Error struct {
			Level string `json:"level"`
			Code  string `json:"code"`
			Text  string `json:"text"`
		} `json:"error"`
		Record struct {
			Total   string `json:"total"`
			Current string `json:"current"`
		} `json:"record"`
	} `json:"response"`
}

// NewVWorldProvider vWorld Provider 생성자
func NewVWorldProvider(apiKey string, httpClient *httpclient.Client, logger *zap.Logger) *VWorldProvider {
	return &VWorldProvider{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    "https://api.vworld.kr/req/address",
		logger:     logger,
	}
}

func (v *VWorldProvider) Name() string {
	return "vWorld"
}

func (v *VWorldProvider) IsAvailable(ctx context.Context) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return !v.disabled
}

// Disable Provider를 비활성화
func (v *VWorldProvider) Disable(reason string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.disabled = true
	v.disableReason = reason
	v.logger.Warn("vWorld provider disabled",
		zap.String("reason", reason),
	)
}

// IsDisabled Provider가 비활성화 되었는지 확인
func (v *VWorldProvider) IsDisabled() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.disabled
}

// GetDisableReason 비활성화 사유 반환
func (v *VWorldProvider) GetDisableReason() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.disableReason
}

func (v *VWorldProvider) Geocode(ctx context.Context, address string) (*model.ProviderResult, error) {
	return v.GeocodeWithType(ctx, address, "")
}

// GeocodeWithType 특정 주소 타입으로 지오코딩 (타입이 빈 문자열이면 자동 폴백)
func (v *VWorldProvider) GeocodeWithType(ctx context.Context, address string, addrType string) (*model.ProviderResult, error) {
	// 주소 전처리
	address = strings.TrimSpace(address)
	if address == "" {
		return &model.ProviderResult{
			Success: false,
			Error:   ErrInvalidAddress,
		}, nil
	}

	// 주소 타입 정규화 (소문자 -> 대문자)
	addrType = strings.ToUpper(addrType)

	// 특정 타입이 지정된 경우 해당 타입만 시도
	if addrType == "ROAD" || addrType == "PARCEL" {
		result, err := v.geocodeWithType(ctx, address, addrType)
		if err != nil {
			return nil, err
		}
		return result, nil
	}

	// 타입이 지정되지 않은 경우 자동 폴백
	// 1단계: 도로명 주소로 시도
	result, err := v.geocodeWithType(ctx, address, "ROAD")
	if err == nil && result.Success {
		v.logger.Debug("vWorld geocoding succeeded with road address",
			zap.String("address", address),
			zap.String("type", "ROAD"),
		)
		return result, nil
	}

	// 2단계: 지번 주소로 재시도
	v.logger.Debug("Retrying with parcel address type",
		zap.String("address", address),
	)
	result, err = v.geocodeWithType(ctx, address, "PARCEL")
	if err == nil && result.Success {
		v.logger.Debug("vWorld geocoding succeeded with parcel address",
			zap.String("address", address),
			zap.String("type", "PARCEL"),
		)
		return result, nil
	}

	// 모두 실패한 경우
	if err != nil {
		return nil, err
	}

	return &model.ProviderResult{
		Success: false,
		Error:   ErrAddressNotFound,
	}, nil
}

func (v *VWorldProvider) geocodeWithType(ctx context.Context, address, addrType string) (*model.ProviderResult, error) {
	// URL 파라미터 구성
	params := url.Values{}
	params.Set("service", "address")
	params.Set("request", "getcoord")
	params.Set("crs", "epsg:4326")     // WGS84 좌표계
	params.Set("address", address)
	params.Set("format", "json")
	params.Set("type", addrType)        // road 또는 parcel
	params.Set("key", v.apiKey)
	
	requestURL := fmt.Sprintf("%s?%s", v.baseURL, params.Encode())
	
	// HTTP 요청 생성
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// HTTP 요청 실행
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, NewClassifiedError(ErrorTypeSystemFailure, "HTTP request failed", err)
	}
	defer resp.Body.Close()
	
	// 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, NewClassifiedError(ErrorTypeUnauthorized, "Invalid API key", ErrAPIKeyInvalid)
		case http.StatusTooManyRequests:
			return nil, NewClassifiedError(ErrorTypeRateLimitExceeded, "Rate limit exceeded", ErrQuotaExceeded)
		default:
			return nil, NewClassifiedError(ErrorTypeSystemFailure, 
				fmt.Sprintf("API returned status %d", resp.StatusCode), nil)
		}
	}
	
	// 응답 파싱
	var vwResp VWorldResponse
	if err := json.NewDecoder(resp.Body).Decode(&vwResp); err != nil {
		return nil, fmt.Errorf("failed to decode vWorld response: %w", err)
	}
	
	// 에러 체크
	if vwResp.Response.Status == "ERROR" {
		errText := vwResp.Response.Error.Text
		v.logger.Warn("vWorld API error",
			zap.String("error_code", vwResp.Response.Error.Code),
			zap.String("error_text", errText),
		)
		
		// 에러 코드에 따른 처리
		if strings.Contains(errText, "인증키") || strings.Contains(errText, "AUTH") {
			return nil, NewClassifiedError(ErrorTypeUnauthorized, errText, nil)
		}
		
		return &model.ProviderResult{
			Success: false,
			Error:   fmt.Errorf("vWorld API error: %s", errText),
		}, nil
	}
	
	// 결과 확인
	if vwResp.Response.Status != "OK" || vwResp.Response.Result.Point.X == "" || vwResp.Response.Result.Point.Y == "" {
		// 실제 API 에러 메시지 사용
		errorMsg := "address not found"
		if vwResp.Response.Status == "NOT_FOUND" {
			errorMsg = "NOT_FOUND: 검색 결과가 없습니다"
		} else if vwResp.Response.Status != "OK" {
			errorMsg = fmt.Sprintf("%s: %s", vwResp.Response.Status, vwResp.Response.Error.Text)
		}

		return &model.ProviderResult{
			Success: false,
			Error:   fmt.Errorf("%s", errorMsg),
		}, nil
	}

	// 좌표 파싱
	lng, err := strconv.ParseFloat(vwResp.Response.Result.Point.X, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}

	lat, err := strconv.ParseFloat(vwResp.Response.Result.Point.Y, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}

	// 주소 정보 추출
	var roadAddr, parcelAddr string
	if vwResp.Response.Input.Type == "ROAD" {
		roadAddr = vwResp.Response.Input.Address
		// refined 정보가 있으면 지번 주소도 구성
		if vwResp.Response.Refined.Text != "" {
			parcelAddr = vwResp.Response.Refined.Text
		}
	} else {
		parcelAddr = vwResp.Response.Input.Address
	}

	v.logger.Info("vWorld geocoding succeeded",
		zap.String("address_type", addrType),
		zap.Float64("latitude", lat),
		zap.Float64("longitude", lng),
	)
	
	return &model.ProviderResult{
		Coordinate: model.Coordinate{
			Latitude:  lat,
			Longitude: lng,
		},
		AddressDetail: model.AddressDetail{
			RoadAddress:   roadAddr,
			ParcelAddress: parcelAddr,
			BuildingName:  vwResp.Response.Refined.Structure.Detail,
		},
		Success: true,
	}, nil
}