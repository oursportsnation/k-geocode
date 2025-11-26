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

package handler

import (
	"net/http"
	"time"
	
	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/oursportsnation/k-geocode/internal/service"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GeocodingHandler 지오코딩 API 핸들러
type GeocodingHandler struct {
	service service.GeocodingServiceInterface
	logger  *zap.Logger
}

// NewGeocodingHandler 지오코딩 핸들러 생성자
func NewGeocodingHandler(service service.GeocodingServiceInterface, logger *zap.Logger) *GeocodingHandler {
	return &GeocodingHandler{
		service: service,
		logger:  logger,
	}
}

// Geocode 단건 지오코딩 API
// @Summary      주소를 좌표로 변환
// @Description  한글 주소를 WGS84 좌표로 변환합니다. vWorld API를 우선 사용하고 실패 시 Kakao API로 자동 폴백됩니다.
// @Description  address_type을 지정하면 해당 타입(ROAD/PARCEL)으로만 검색합니다. 미지정 시 자동으로 ROAD → PARCEL 순서로 시도합니다.
// @Tags         geocoding
// @Accept       json
// @Produce      json
// @Param        request body model.GeocodingRequest true "지오코딩 요청 (address_type은 선택사항: ROAD 또는 PARCEL)"
// @Success      200 {object} model.GeocodingResponse "변환 성공"
// @Success      404 {object} model.GeocodingResponse "주소를 찾을 수 없음"
// @Failure      400 {object} map[string]string "잘못된 요청"
// @Failure      500 {object} map[string]string "서버 에러"
// @Router       /api/v1/geocode [post]
func (h *GeocodingHandler) Geocode(c *gin.Context) {
	start := time.Now()
	
	// Request ID 가져오기 (미들웨어에서 설정)
	requestID := c.GetString("requestID")
	
	// 요청 파싱
	var req model.GeocodingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request format",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}
	
	h.logger.Info("Geocoding request received",
		zap.String("request_id", requestID),
		zap.String("address", req.Address),
		zap.String("address_type", req.AddressType),
	)

	// 지오코딩 서비스 호출
	resp, err := h.service.Geocode(c.Request.Context(), req.Address, req.AddressType)
	if err != nil {
		h.logger.Error("Geocoding service error",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	
	// 응답 시간 로깅
	h.logger.Info("Geocoding request completed",
		zap.String("request_id", requestID),
		zap.Bool("success", resp.Success),
		zap.String("provider", resp.Provider),
		zap.Duration("duration", time.Since(start)),
	)
	
	// 성공/실패에 따른 상태 코드 설정
	statusCode := http.StatusOK
	if !resp.Success {
		statusCode = http.StatusNotFound
	}
	
	c.JSON(statusCode, resp)
}

// GeocodeBulk 대량 지오코딩 API
// @Summary      여러 주소를 좌표로 변환
// @Description  여러 한글 주소를 WGS84 좌표로 변환합니다. 최대 100개까지 처리 가능하며, 최대 10개씩 동시 처리됩니다.
// @Tags         geocoding
// @Accept       json
// @Produce      json
// @Param        request body model.BulkRequest true "대량 지오코딩 요청 (최대 100개)"
// @Success      200 {object} model.BulkResponse "변환 결과"
// @Failure      400 {object} map[string]string "잘못된 요청 (빈 배열 또는 100개 초과)"
// @Failure      500 {object} map[string]string "서버 에러"
// @Router       /api/v1/geocode/bulk [post]
func (h *GeocodingHandler) GeocodeBulk(c *gin.Context) {
	start := time.Now()
	requestID := c.GetString("requestID")
	
	// 요청 파싱
	var req model.BulkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid bulk request format",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request format",
		})
		return
	}
	
	// 최대 개수 검증
	if len(req.Addresses) > 100 {
		h.logger.Warn("Too many addresses in bulk request",
			zap.String("request_id", requestID),
			zap.Int("count", len(req.Addresses)),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "maximum 100 addresses allowed",
		})
		return
	}
	
	h.logger.Info("Bulk geocoding request received",
		zap.String("request_id", requestID),
		zap.Int("address_count", len(req.Addresses)),
	)
	
	// 배치 지오코딩 서비스 호출
	resp, err := h.service.GeocodeBatch(c.Request.Context(), req.Addresses)
	if err != nil {
		h.logger.Error("Bulk geocoding service error",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}
	
	h.logger.Info("Bulk geocoding request completed",
		zap.String("request_id", requestID),
		zap.Int("total", resp.Summary.Total),
		zap.Int("success", resp.Summary.Success),
		zap.Int("failed", resp.Summary.Failed),
		zap.Duration("duration", time.Since(start)),
	)
	
	c.JSON(http.StatusOK, resp)
}