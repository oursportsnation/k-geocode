package unit

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oursportsnation/k-geocode/internal/handler"
	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/oursportsnation/k-geocode/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockGeocodingService 지오코딩 서비스 모킹
type MockGeocodingService struct {
	mock.Mock
}

// Geocode implements service.GeocodingServiceInterface
func (m *MockGeocodingService) Geocode(ctx context.Context, address string, addressType string) (*model.GeocodingResponse, error) {
	args := m.Called(ctx, address, addressType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.GeocodingResponse), args.Error(1)
}

// GeocodeBatch implements service.GeocodingServiceInterface
func (m *MockGeocodingService) GeocodeBatch(ctx context.Context, addresses []string) (*model.BulkResponse, error) {
	args := m.Called(ctx, addresses)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.BulkResponse), args.Error(1)
}

// MockCoordinator 코디네이터 모킹
type MockCoordinator struct {
	mock.Mock
}

// HealthCheck implements service.CoordinatorInterface
func (m *MockCoordinator) HealthCheck(ctx context.Context) service.HealthStatus {
	args := m.Called(ctx)
	return *args.Get(0).(*service.HealthStatus)
}

// GetGeocodingService implements service.CoordinatorInterface
func (m *MockCoordinator) GetGeocodingService() *service.GeocodingService {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*service.GeocodingService)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestGeocodingHandler_Geocode_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockService := new(MockGeocodingService)

	// Mock 응답 설정
	expectedResp := &model.GeocodingResponse{
		Success:  true,
		Provider: "vworld",
		Coordinate: &model.Coordinate{
			Latitude:  37.123456,
			Longitude: 127.123456,
		},
	}
	mockService.On("Geocode", mock.Anything, "서울시 강남구 테헤란로 123", mock.Anything).Return(expectedResp, nil)

	// Handler 생성
	h := handler.NewGeocodingHandler(mockService, logger)
	router.POST("/geocode", h.Geocode)

	// 요청 생성
	reqBody := model.GeocodingRequest{Address: "서울시 강남구 테헤란로 123"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/geocode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusOK, w.Code)

	var resp model.GeocodingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "vworld", resp.Provider)
	assert.NotNil(t, resp.Coordinate)
	assert.Equal(t, 37.123456, resp.Coordinate.Latitude)

	mockService.AssertExpectations(t)
}

func TestGeocodingHandler_Geocode_NotFound(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockService := new(MockGeocodingService)

	// Mock 응답 설정 - 주소를 찾지 못함
	expectedResp := &model.GeocodingResponse{
		Success: false,
		Error:   "address not found",
	}
	mockService.On("Geocode", mock.Anything, "존재하지않는주소123456", mock.Anything).Return(expectedResp, nil)

	// Handler 생성
	h := handler.NewGeocodingHandler(mockService, logger)
	router.POST("/geocode", h.Geocode)

	// 요청 생성
	reqBody := model.GeocodingRequest{Address: "존재하지않는주소123456"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/geocode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp model.GeocodingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "address not found", resp.Error)

	mockService.AssertExpectations(t)
}

func TestGeocodingHandler_Geocode_InvalidRequest(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockService := new(MockGeocodingService)

	// Handler 생성
	h := handler.NewGeocodingHandler(mockService, logger)
	router.POST("/geocode", h.Geocode)

	// 잘못된 JSON 요청
	req := httptest.NewRequest(http.MethodPost, "/geocode", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "invalid request format")
}

func TestGeocodingHandler_Geocode_ServiceError(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockService := new(MockGeocodingService)

	// Mock 응답 설정 - 서비스 에러
	mockService.On("Geocode", mock.Anything, "서울시 강남구", mock.Anything).Return(nil, assert.AnError)

	// Handler 생성
	h := handler.NewGeocodingHandler(mockService, logger)
	router.POST("/geocode", h.Geocode)

	// 요청 생성
	reqBody := model.GeocodingRequest{Address: "서울시 강남구"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/geocode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "internal server error", resp["error"])

	mockService.AssertExpectations(t)
}

func TestGeocodingHandler_GeocodeBulk_Success(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockService := new(MockGeocodingService)

	// Mock 응답 설정
	addresses := []string{"서울시 강남구", "부산시 해운대구"}
	expectedResp := &model.BulkResponse{
		Results: []*model.GeocodingResponse{
			{
				Success:  true,
				Provider: "vworld",
				Coordinate: &model.Coordinate{
					Latitude:  37.123456,
					Longitude: 127.123456,
				},
			},
			{
				Success:  true,
				Provider: "kakao",
				Coordinate: &model.Coordinate{
					Latitude:  35.123456,
					Longitude: 129.123456,
				},
			},
		},
	}
	expectedResp.Summary.Total = 2
	expectedResp.Summary.Success = 2
	expectedResp.Summary.Failed = 0
	mockService.On("GeocodeBatch", mock.Anything, addresses).Return(expectedResp, nil)

	// Handler 생성
	h := handler.NewGeocodingHandler(mockService, logger)
	router.POST("/geocode/bulk", h.GeocodeBulk)

	// 요청 생성
	reqBody := model.BulkRequest{Addresses: addresses}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/geocode/bulk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusOK, w.Code)

	var resp model.BulkResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Summary.Total)
	assert.Equal(t, 2, resp.Summary.Success)
	assert.Equal(t, 0, resp.Summary.Failed)
	assert.Len(t, resp.Results, 2)

	mockService.AssertExpectations(t)
}

func TestGeocodingHandler_GeocodeBulk_TooManyAddresses(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockService := new(MockGeocodingService)

	// Handler 생성
	h := handler.NewGeocodingHandler(mockService, logger)
	router.POST("/geocode/bulk", h.GeocodeBulk)

	// 101개 주소 생성 (최대 100개)
	addresses := make([]string, 101)
	for i := 0; i < 101; i++ {
		addresses[i] = "서울시 강남구"
	}

	// 요청 생성
	reqBody := model.BulkRequest{Addresses: addresses}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/geocode/bulk", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	// Gin의 binding validation이 먼저 실패하므로 "invalid request format" 에러 발생
	assert.Equal(t, "invalid request format", resp["error"])
}

func TestGeocodingHandler_GeocodeBulk_InvalidRequest(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockService := new(MockGeocodingService)

	// Handler 생성
	h := handler.NewGeocodingHandler(mockService, logger)
	router.POST("/geocode/bulk", h.GeocodeBulk)

	// 잘못된 JSON 요청
	req := httptest.NewRequest(http.MethodPost, "/geocode/bulk", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "invalid request format")
}

func TestHealthHandler_Ping(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockCoordinator := new(MockCoordinator)

	// Handler 생성
	h := handler.NewHealthHandler(mockCoordinator, logger)
	router.GET("/ping", h.Ping)

	// 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "pong", resp["message"])
	assert.NotEmpty(t, resp["time"])
}

func TestHealthHandler_Health_Healthy(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockCoordinator := new(MockCoordinator)

	// Mock 응답 설정 - 모든 Provider 정상
	healthStatus := &service.HealthStatus{
		Healthy: true,
		Providers: []service.ProviderStatus{
			{Name: "vworld", Available: true},
			{Name: "kakao", Available: true},
		},
	}
	mockCoordinator.On("HealthCheck", mock.Anything).Return(healthStatus)

	// Handler 생성
	h := handler.NewHealthHandler(mockCoordinator, logger)
	router.GET("/health", h.Health)

	// 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", resp["status"])
	assert.NotNil(t, resp["system"])
	assert.NotNil(t, resp["providers"])

	mockCoordinator.AssertExpectations(t)
}

func TestHealthHandler_Health_Unhealthy(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockCoordinator := new(MockCoordinator)

	// Mock 응답 설정 - 일부 Provider 장애
	healthStatus := &service.HealthStatus{
		Healthy: false,
		Providers: []service.ProviderStatus{
			{Name: "vworld", Available: false},
			{Name: "kakao", Available: true},
		},
	}
	mockCoordinator.On("HealthCheck", mock.Anything).Return(healthStatus)

	// Handler 생성
	h := handler.NewHealthHandler(mockCoordinator, logger)
	router.GET("/health", h.Health)

	// 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unhealthy", resp["status"])

	mockCoordinator.AssertExpectations(t)
}

func TestHealthHandler_Ready_Ready(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockCoordinator := new(MockCoordinator)

	// Mock 응답 설정 - Ready 상태
	healthStatus := &service.HealthStatus{
		Healthy: true,
		Providers: []service.ProviderStatus{
			{Name: "vworld", Available: true},
			{Name: "kakao", Available: true},
		},
	}
	mockCoordinator.On("HealthCheck", mock.Anything).Return(healthStatus)

	// Handler 생성
	h := handler.NewHealthHandler(mockCoordinator, logger)
	router.GET("/ready", h.Ready)

	// 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp["ready"])

	mockCoordinator.AssertExpectations(t)
}

func TestHealthHandler_Ready_NotReady(t *testing.T) {
	// Setup
	router := setupTestRouter()
	logger := zap.NewNop()
	mockCoordinator := new(MockCoordinator)

	// Mock 응답 설정 - Not Ready 상태
	healthStatus := &service.HealthStatus{
		Healthy: false,
		Providers: []service.ProviderStatus{
			{Name: "vworld", Available: false},
			{Name: "kakao", Available: false},
		},
	}
	mockCoordinator.On("HealthCheck", mock.Anything).Return(healthStatus)

	// Handler 생성
	h := handler.NewHealthHandler(mockCoordinator, logger)
	router.GET("/ready", h.Ready)

	// 요청 생성
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	// 실행
	router.ServeHTTP(w, req)

	// 검증
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp["ready"])

	mockCoordinator.AssertExpectations(t)
}
