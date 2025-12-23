package service

import (
	"context"
	"errors"
	"testing"

	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/oursportsnation/k-geocode/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockProvider is a test mock for GeocodingProvider
type mockProvider struct {
	name           string
	available      bool
	disabled       bool
	disableReason  string
	result         *model.ProviderResult
	err            error
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) IsAvailable(ctx context.Context) bool { return m.available && !m.disabled }
func (m *mockProvider) Disable(reason string) { m.disabled = true; m.disableReason = reason }
func (m *mockProvider) IsDisabled() bool { return m.disabled }
func (m *mockProvider) GetDisableReason() string { return m.disableReason }
func (m *mockProvider) Geocode(ctx context.Context, address string) (*model.ProviderResult, error) {
	return m.result, m.err
}

func TestNewGeocodingService(t *testing.T) {
	logger := zap.NewNop()
	providers := []provider.GeocodingProvider{
		&mockProvider{name: "mock1", available: true},
	}

	svc := NewGeocodingService(providers, logger)

	require.NotNil(t, svc)
	assert.Equal(t, providers, svc.providers)
	assert.Equal(t, logger, svc.logger)
}

func TestGeocodingService_Geocode_Success(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{
		name:      "MockProvider",
		available: true,
		result: &model.ProviderResult{
			Success: true,
			Coordinate: model.Coordinate{
				Latitude:  37.5665,
				Longitude: 126.978,
			},
			AddressDetail: model.AddressDetail{
				RoadAddress: "서울특별시 중구 세종대로 110",
			},
		},
	}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP}, logger)

	result, err := svc.Geocode(context.Background(), "서울특별시 중구 세종대로 110", "")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "MockProvider", result.Provider)
	assert.InDelta(t, 37.5665, result.Coordinate.Latitude, 0.0001)
	assert.InDelta(t, 126.978, result.Coordinate.Longitude, 0.0001)
}

func TestGeocodingService_Geocode_InvalidAddress(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{name: "MockProvider", available: true}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP}, logger)

	result, err := svc.Geocode(context.Background(), "ab", "") // Too short

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "invalid address")
}

func TestGeocodingService_Geocode_ProviderNotAvailable(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{
		name:      "MockProvider",
		available: false,
	}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP}, logger)

	result, err := svc.Geocode(context.Background(), "서울특별시 중구 세종대로", "")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "none", result.Provider)
}

func TestGeocodingService_Geocode_Fallback(t *testing.T) {
	logger := zap.NewNop()
	failingProvider := &mockProvider{
		name:      "FailingProvider",
		available: true,
		result:    &model.ProviderResult{Success: false},
	}
	successProvider := &mockProvider{
		name:      "SuccessProvider",
		available: true,
		result: &model.ProviderResult{
			Success: true,
			Coordinate: model.Coordinate{
				Latitude:  37.5665,
				Longitude: 126.978,
			},
		},
	}
	svc := NewGeocodingService([]provider.GeocodingProvider{failingProvider, successProvider}, logger)

	result, err := svc.Geocode(context.Background(), "서울특별시 중구 세종대로", "")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "SuccessProvider", result.Provider)
	assert.Len(t, result.Attempts, 2)
}

func TestGeocodingService_Geocode_ClassifiedError(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{
		name:      "MockProvider",
		available: true,
		err:       provider.NewClassifiedError(provider.ErrorTypeNotFound, "not found", nil),
	}
	backupProvider := &mockProvider{
		name:      "BackupProvider",
		available: true,
		result: &model.ProviderResult{
			Success: true,
			Coordinate: model.Coordinate{
				Latitude:  37.5665,
				Longitude: 126.978,
			},
		},
	}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP, backupProvider}, logger)

	result, err := svc.Geocode(context.Background(), "서울특별시 중구 세종대로", "")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "BackupProvider", result.Provider)
}

func TestGeocodingService_Geocode_UnauthorizedDisablesProvider(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{
		name:      "MockProvider",
		available: true,
		err:       provider.NewClassifiedError(provider.ErrorTypeUnauthorized, "auth failed", nil),
	}
	backupProvider := &mockProvider{
		name:      "BackupProvider",
		available: true,
		result: &model.ProviderResult{
			Success: true,
			Coordinate: model.Coordinate{
				Latitude:  37.5665,
				Longitude: 126.978,
			},
		},
	}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP, backupProvider}, logger)

	result, err := svc.Geocode(context.Background(), "서울특별시 중구 세종대로", "")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "BackupProvider", result.Provider)
	assert.True(t, mockP.IsDisabled())
}

func TestGeocodingService_Geocode_NonFallbackError(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{
		name:      "MockProvider",
		available: true,
		err:       provider.NewClassifiedError(provider.ErrorTypeInvalid, "invalid input", nil),
	}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP}, logger)

	result, err := svc.Geocode(context.Background(), "서울특별시 중구 세종대로", "")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "invalid input")
}

func TestGeocodingService_Geocode_UnexpectedError(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{
		name:      "MockProvider",
		available: true,
		err:       errors.New("unexpected error"),
	}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP}, logger)

	result, err := svc.Geocode(context.Background(), "서울특별시 중구 세종대로", "")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Success)
}

func TestGeocodingService_GeocodeBatch_Success(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{
		name:      "MockProvider",
		available: true,
		result: &model.ProviderResult{
			Success: true,
			Coordinate: model.Coordinate{
				Latitude:  37.5665,
				Longitude: 126.978,
			},
		},
	}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP}, logger)

	addresses := []string{
		"서울특별시 중구 세종대로 110",
		"부산광역시 해운대구 해운대해변로 264",
	}
	result, err := svc.GeocodeBatch(context.Background(), addresses)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 2, result.Summary.Total)
	assert.Equal(t, 2, result.Summary.Success)
	assert.Equal(t, 0, result.Summary.Failed)
	assert.Len(t, result.Results, 2)
}

func TestGeocodingService_GeocodeBatch_Empty(t *testing.T) {
	logger := zap.NewNop()
	mockP := &mockProvider{name: "MockProvider", available: true}
	svc := NewGeocodingService([]provider.GeocodingProvider{mockP}, logger)

	result, err := svc.GeocodeBatch(context.Background(), []string{})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Results)
}

func TestGeocodingService_ValidateAddress(t *testing.T) {
	logger := zap.NewNop()
	svc := NewGeocodingService(nil, logger)

	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"valid address", "서울특별시 중구 세종대로 110", false},
		{"invalid short address", "ab", true},
		{"empty address", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateAddress(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGeocodingService_GetAvailableProviders(t *testing.T) {
	logger := zap.NewNop()
	providers := []provider.GeocodingProvider{
		&mockProvider{name: "Provider1", available: true},
		&mockProvider{name: "Provider2", available: false},
		&mockProvider{name: "Provider3", available: true},
	}
	svc := NewGeocodingService(providers, logger)

	result := svc.GetAvailableProviders(context.Background())

	assert.Len(t, result, 2)
	assert.Contains(t, result, "Provider1")
	assert.Contains(t, result, "Provider3")
	assert.NotContains(t, result, "Provider2")
}
