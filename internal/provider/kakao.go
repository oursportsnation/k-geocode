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
	
	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/oursportsnation/k-geocode/pkg/httpclient"
	
	"go.uber.org/zap"
)

// KakaoProvider Kakao Local API 클라이언트
type KakaoProvider struct {
	apiKey     string
	httpClient *httpclient.Client
	baseURL    string
	logger     *zap.Logger
}

// KakaoResponse Kakao API 응답 구조체
type KakaoResponse struct {
	Meta struct {
		TotalCount    int  `json:"total_count"`
		PageableCount int  `json:"pageable_count"`
		IsEnd         bool `json:"is_end"`
	} `json:"meta"`
	Documents []struct {
		AddressName string `json:"address_name"`
		X           string `json:"x"` // 경도
		Y           string `json:"y"` // 위도
		AddressType string `json:"address_type"` // REGION(지명), ROAD(도로명), REGION_ADDR(지번)
		Address     struct {
			AddressName       string `json:"address_name"`
			Region1depthName  string `json:"region_1depth_name"`
			Region2depthName  string `json:"region_2depth_name"`
			Region3depthName  string `json:"region_3depth_name"`
			Region3depthHName string `json:"region_3depth_h_name"`
			HCode             string `json:"h_code"`
			BCode             string `json:"b_code"`
			MountainYn        string `json:"mountain_yn"`
			MainAddressNo     string `json:"main_address_no"`
			SubAddressNo      string `json:"sub_address_no"`
		} `json:"address"`
		RoadAddress struct {
			AddressName       string `json:"address_name"`
			Region1depthName  string `json:"region_1depth_name"`
			Region2depthName  string `json:"region_2depth_name"`
			Region3depthName  string `json:"region_3depth_name"`
			RoadName          string `json:"road_name"`
			UndergroundYn     string `json:"underground_yn"`
			MainBuildingNo    string `json:"main_building_no"`
			SubBuildingNo     string `json:"sub_building_no"`
			BuildingName      string `json:"building_name"`
			ZoneNo            string `json:"zone_no"` // 우편번호
		} `json:"road_address"`
	} `json:"documents"`
}

// KakaoErrorResponse Kakao API 에러 응답
type KakaoErrorResponse struct {
	ErrorType string `json:"errorType"`
	Message   string `json:"message"`
}

// NewKakaoProvider Kakao Provider 생성자
func NewKakaoProvider(apiKey string, httpClient *httpclient.Client, logger *zap.Logger) *KakaoProvider {
	return &KakaoProvider{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    "https://dapi.kakao.com/v2/local/search/address.json",
		logger:     logger,
	}
}

func (k *KakaoProvider) Name() string {
	return "Kakao"
}

func (k *KakaoProvider) IsAvailable(ctx context.Context) bool {
	// 기본적으로 사용 가능
	// 추후 Circuit Breaker 통합 시 상태 확인 추가
	return true
}

func (k *KakaoProvider) Geocode(ctx context.Context, address string) (*model.ProviderResult, error) {
	// 주소 전처리
	address = strings.TrimSpace(address)
	if address == "" {
		return &model.ProviderResult{
			Success: false,
			Error:   ErrInvalidAddress,
		}, nil
	}
	
	// URL 파라미터
	params := url.Values{}
	params.Set("query", address)
	params.Set("analyze_type", "similar") // similar 또는 exact
	params.Set("size", "10")              // 최대 10개 결과
	
	requestURL := fmt.Sprintf("%s?%s", k.baseURL, params.Encode())
	
	// HTTP 요청 생성
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Kakao API 인증 헤더
	req.Header.Set("Authorization", fmt.Sprintf("KakaoAK %s", k.apiKey))
	
	// HTTP 요청 실행
	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, NewClassifiedError(ErrorTypeSystemFailure, "HTTP request failed", err)
	}
	defer resp.Body.Close()
	
	// 상태 코드 확인
	if resp.StatusCode != http.StatusOK {
		// 에러 응답 파싱 시도
		var errResp KakaoErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			k.logger.Warn("Kakao API error response",
				zap.String("error_type", errResp.ErrorType),
				zap.String("message", errResp.Message),
			)
		}
		
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, NewClassifiedError(ErrorTypeUnauthorized, "Invalid API key", ErrAPIKeyInvalid)
		case http.StatusBadRequest:
			return nil, NewClassifiedError(ErrorTypeInvalid, "Bad request", nil)
		case http.StatusTooManyRequests:
			return nil, NewClassifiedError(ErrorTypeRateLimitExceeded, "Rate limit exceeded", ErrQuotaExceeded)
		default:
			return nil, NewClassifiedError(ErrorTypeSystemFailure,
				fmt.Sprintf("API returned status %d", resp.StatusCode), nil)
		}
	}
	
	// 응답 파싱
	var kakaoResp KakaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&kakaoResp); err != nil {
		return nil, fmt.Errorf("failed to decode Kakao response: %w", err)
	}
	
	// 결과 없음
	if len(kakaoResp.Documents) == 0 {
		k.logger.Debug("Kakao returned no results",
			zap.String("address", address),
			zap.Int("total_count", kakaoResp.Meta.TotalCount),
		)
		return &model.ProviderResult{
			Success: false,
			Error:   ErrAddressNotFound,
		}, nil
	}
	
	// 첫 번째 결과 사용
	doc := kakaoResp.Documents[0]
	
	// 좌표 파싱
	lng, err := strconv.ParseFloat(doc.X, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}
	
	lat, err := strconv.ParseFloat(doc.Y, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}
	
	// 주소 정보 구성
	var roadAddr, parcelAddr, zipcode, buildingName string
	
	// 도로명 주소 정보가 있는 경우
	if doc.RoadAddress.AddressName != "" {
		roadAddr = doc.RoadAddress.AddressName
		zipcode = doc.RoadAddress.ZoneNo
		buildingName = doc.RoadAddress.BuildingName
	}
	
	// 지번 주소 정보
	if doc.Address.AddressName != "" {
		parcelAddr = doc.Address.AddressName
	}
	
	// 도로명 주소가 없고 지번 주소만 있는 경우
	if roadAddr == "" && parcelAddr != "" {
		// 일부 경우 address_name에 전체 주소가 들어있음
		if doc.AddressType == "ROAD" {
			roadAddr = doc.AddressName
		} else {
			parcelAddr = doc.AddressName
		}
	}
	
	k.logger.Info("Kakao geocoding succeeded",
		zap.Float64("latitude", lat),
		zap.Float64("longitude", lng),
		zap.String("address_type", doc.AddressType),
		zap.Int("total_results", kakaoResp.Meta.TotalCount),
	)
	
	return &model.ProviderResult{
		Coordinate: model.Coordinate{
			Latitude:  lat,
			Longitude: lng,
		},
		AddressDetail: model.AddressDetail{
			RoadAddress:   roadAddr,
			ParcelAddress: parcelAddr,
			Zipcode:       zipcode,
			BuildingName:  buildingName,
		},
		Success: true,
	}, nil
}