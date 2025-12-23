package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockGeocodingService implements service.GeocodingServiceInterface for testing
type mockGeocodingService struct {
	geocodeResult *model.GeocodingResponse
	geocodeErr    error
	batchResult   *model.BulkResponse
	batchErr      error
}

func (m *mockGeocodingService) Geocode(ctx context.Context, address string, addressType string) (*model.GeocodingResponse, error) {
	return m.geocodeResult, m.geocodeErr
}

func (m *mockGeocodingService) GeocodeBatch(ctx context.Context, addresses []string) (*model.BulkResponse, error) {
	return m.batchResult, m.batchErr
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewGeocodingHandler(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{}

	handler := NewGeocodingHandler(mockService, logger)

	require.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
	assert.Equal(t, logger, handler.logger)
}

func TestGeocodingHandler_Geocode_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{
		geocodeResult: &model.GeocodingResponse{
			Success:  true,
			Provider: "vWorld",
			Coordinate: &model.Coordinate{
				Latitude:  37.5665,
				Longitude: 126.978,
			},
		},
	}
	handler := NewGeocodingHandler(mockService, logger)

	router := setupTestRouter()
	router.POST("/geocode", handler.Geocode)

	body := `{"address": "서울특별시 중구 세종대로 110"}`
	req := httptest.NewRequest(http.MethodPost, "/geocode", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp model.GeocodingResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "vWorld", resp.Provider)
}

func TestGeocodingHandler_Geocode_NotFound(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{
		geocodeResult: &model.GeocodingResponse{
			Success:  false,
			Provider: "none",
			Error:    "address not found",
		},
	}
	handler := NewGeocodingHandler(mockService, logger)

	router := setupTestRouter()
	router.POST("/geocode", handler.Geocode)

	body := `{"address": "없는 주소"}`
	req := httptest.NewRequest(http.MethodPost, "/geocode", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGeocodingHandler_Geocode_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{}
	handler := NewGeocodingHandler(mockService, logger)

	router := setupTestRouter()
	router.POST("/geocode", handler.Geocode)

	// Missing required "address" field
	body := `{"wrong_field": "value"}`
	req := httptest.NewRequest(http.MethodPost, "/geocode", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGeocodingHandler_Geocode_ServiceError(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{
		geocodeErr: errors.New("service error"),
	}
	handler := NewGeocodingHandler(mockService, logger)

	router := setupTestRouter()
	router.POST("/geocode", handler.Geocode)

	body := `{"address": "서울시"}`
	req := httptest.NewRequest(http.MethodPost, "/geocode", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGeocodingHandler_GeocodeBulk_Success(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{
		batchResult: &model.BulkResponse{
			Results: []*model.GeocodingResponse{
				{Success: true, Provider: "vWorld"},
				{Success: true, Provider: "vWorld"},
			},
			Summary: struct {
				Total   int `json:"total"`
				Success int `json:"success"`
				Failed  int `json:"failed"`
			}{Total: 2, Success: 2, Failed: 0},
		},
	}
	handler := NewGeocodingHandler(mockService, logger)

	router := setupTestRouter()
	router.POST("/geocode/bulk", handler.GeocodeBulk)

	body := `{"addresses": ["서울시 중구", "부산시 해운대구"]}`
	req := httptest.NewRequest(http.MethodPost, "/geocode/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp model.BulkResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Summary.Total)
}

func TestGeocodingHandler_GeocodeBulk_TooManyAddresses(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{}
	handler := NewGeocodingHandler(mockService, logger)

	router := setupTestRouter()
	router.POST("/geocode/bulk", handler.GeocodeBulk)

	// Create more than 100 addresses
	addresses := make([]string, 101)
	for i := range addresses {
		addresses[i] = "서울시"
	}
	bodyBytes, _ := json.Marshal(map[string][]string{"addresses": addresses})
	req := httptest.NewRequest(http.MethodPost, "/geocode/bulk", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGeocodingHandler_GeocodeBulk_InvalidRequest(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{}
	handler := NewGeocodingHandler(mockService, logger)

	router := setupTestRouter()
	router.POST("/geocode/bulk", handler.GeocodeBulk)

	body := `{"wrong_field": "value"}`
	req := httptest.NewRequest(http.MethodPost, "/geocode/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGeocodingHandler_GeocodeBulk_ServiceError(t *testing.T) {
	logger := zap.NewNop()
	mockService := &mockGeocodingService{
		batchErr: errors.New("service error"),
	}
	handler := NewGeocodingHandler(mockService, logger)

	router := setupTestRouter()
	router.POST("/geocode/bulk", handler.GeocodeBulk)

	body := `{"addresses": ["서울시"]}`
	req := httptest.NewRequest(http.MethodPost, "/geocode/bulk", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
