package handler

import (
	"net/http"
	"runtime"
	"time"
	
	"github.com/oursportsnation/k-geocode/internal/service"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthHandler 헬스체크 API 핸들러
type HealthHandler struct {
	coordinator service.CoordinatorInterface
	logger      *zap.Logger
	startTime   time.Time
}

// NewHealthHandler 헬스 핸들러 생성자
func NewHealthHandler(coordinator service.CoordinatorInterface, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		coordinator: coordinator,
		logger:      logger,
		startTime:   time.Now(),
	}
}

// Health 헬스체크 API
// @Summary      서비스 상태 확인
// @Description  서비스와 Provider들의 상태를 확인합니다. 시스템 정보(메모리, Goroutine 등)도 함께 제공됩니다.
// @Tags         health
// @Produce      json
// @Success      200 {object} HealthResponse "서비스 정상"
// @Success      503 {object} HealthResponse "서비스 비정상 (Provider 장애)"
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	// 시스템 헬스 체크
	healthStatus := h.coordinator.HealthCheck(c.Request.Context())
	
	// 시스템 정보
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Providers: make([]ProviderStatus, 0),
		System: SystemInfo{
			Uptime:      time.Since(h.startTime).String(),
			Goroutines:  runtime.NumGoroutine(),
			MemoryMB:    float64(m.Alloc) / 1024 / 1024,
			NumGC:       m.NumGC,
		},
	}
	
	// Provider 상태 추가
	for _, ps := range healthStatus.Providers {
		response.Providers = append(response.Providers, ProviderStatus{
			Name:      ps.Name,
			Available: ps.Available,
		})
	}
	
	// 전체 상태 설정
	if !healthStatus.Healthy {
		response.Status = "unhealthy"
	}
	
	// 상태에 따른 HTTP 코드
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
		h.logger.Warn("Health check failed",
			zap.String("status", response.Status),
			zap.Any("providers", response.Providers),
		)
	}
	
	c.JSON(statusCode, response)
}

// Ping 간단한 ping 체크
// @Summary      Ping
// @Description  서비스가 살아있는지 간단히 확인합니다
// @Tags         health
// @Produce      json
// @Success      200 {object} map[string]string "pong 응답"
// @Router       /ping [get]
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// Ready readiness 체크
// @Summary      Readiness 체크
// @Description  서비스가 요청을 처리할 준비가 되었는지 확인합니다. Kubernetes Readiness Probe에 사용할 수 있습니다.
// @Tags         health
// @Produce      json
// @Success      200 {object} map[string]bool "준비 완료"
// @Success      503 {object} map[string]bool "준비 안됨"
// @Router       /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	healthStatus := h.coordinator.HealthCheck(c.Request.Context())
	
	ready := healthStatus.Healthy
	statusCode := http.StatusOK
	
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}
	
	c.JSON(statusCode, gin.H{
		"ready": ready,
	})
}

// HealthResponse 헬스체크 응답
type HealthResponse struct {
	Status    string           `json:"status"`
	Timestamp time.Time        `json:"timestamp"`
	Providers []ProviderStatus `json:"providers"`
	System    SystemInfo       `json:"system"`
}

// ProviderStatus Provider 상태
type ProviderStatus struct {
	Name      string `json:"name"`
	Available bool   `json:"available"`
}

// SystemInfo 시스템 정보
type SystemInfo struct {
	Uptime     string  `json:"uptime"`
	Goroutines int     `json:"goroutines"`
	MemoryMB   float64 `json:"memory_mb"`
	NumGC      uint32  `json:"num_gc"`
}