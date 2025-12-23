package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oursportsnation/k-geocode/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockCoordinator implements service.CoordinatorInterface for testing
type mockCoordinator struct {
	healthStatus service.HealthStatus
}

func (m *mockCoordinator) HealthCheck(ctx context.Context) service.HealthStatus {
	return m.healthStatus
}

func (m *mockCoordinator) GetGeocodingService() *service.GeocodingService {
	return nil
}

func TestNewHealthHandler(t *testing.T) {
	logger := zap.NewNop()
	mockCoord := &mockCoordinator{}

	handler := NewHealthHandler(mockCoord, logger)

	require.NotNil(t, handler)
	assert.Equal(t, mockCoord, handler.coordinator)
	assert.Equal(t, logger, handler.logger)
	assert.False(t, handler.startTime.IsZero())
}

func TestHealthHandler_Ping(t *testing.T) {
	logger := zap.NewNop()
	mockCoord := &mockCoordinator{}
	handler := NewHealthHandler(mockCoord, logger)

	router := setupTestRouter()
	router.GET("/ping", handler.Ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "pong", resp["message"])
	assert.NotEmpty(t, resp["time"])
}

func TestHealthHandler_Health_Healthy(t *testing.T) {
	logger := zap.NewNop()
	mockCoord := &mockCoordinator{
		healthStatus: service.HealthStatus{
			Healthy: true,
			Providers: []service.ProviderStatus{
				{Name: "vWorld", Available: true},
				{Name: "Kakao", Available: true},
			},
		},
	}
	handler := NewHealthHandler(mockCoord, logger)

	router := setupTestRouter()
	router.GET("/health", handler.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp.Status)
	assert.Len(t, resp.Providers, 2)
}

func TestHealthHandler_Health_Unhealthy(t *testing.T) {
	logger := zap.NewNop()
	mockCoord := &mockCoordinator{
		healthStatus: service.HealthStatus{
			Healthy: false,
			Providers: []service.ProviderStatus{
				{Name: "vWorld", Available: false},
				{Name: "Kakao", Available: false},
			},
		},
	}
	handler := NewHealthHandler(mockCoord, logger)

	router := setupTestRouter()
	router.GET("/health", handler.Health)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unhealthy", resp.Status)
}

func TestHealthHandler_Ready_Ready(t *testing.T) {
	logger := zap.NewNop()
	mockCoord := &mockCoordinator{
		healthStatus: service.HealthStatus{
			Healthy: true,
		},
	}
	handler := NewHealthHandler(mockCoord, logger)

	router := setupTestRouter()
	router.GET("/ready", handler.Ready)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["ready"])
}

func TestHealthHandler_Ready_NotReady(t *testing.T) {
	logger := zap.NewNop()
	mockCoord := &mockCoordinator{
		healthStatus: service.HealthStatus{
			Healthy: false,
		},
	}
	handler := NewHealthHandler(mockCoord, logger)

	router := setupTestRouter()
	router.GET("/ready", handler.Ready)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]bool
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["ready"])
}
