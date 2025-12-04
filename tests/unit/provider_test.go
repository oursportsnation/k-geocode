package unit

import (
	"context"
	"testing"
	
	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/oursportsnation/k-geocode/internal/provider"
)

// MockProvider 테스트용 Mock Provider
type MockProvider struct {
	name          string
	success       bool
	available     bool
	result        *model.ProviderResult
	disabled      bool
	disableReason string
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) IsAvailable(ctx context.Context) bool {
	return m.available && !m.disabled
}

func (m *MockProvider) Disable(reason string) {
	m.disabled = true
	m.disableReason = reason
}

func (m *MockProvider) IsDisabled() bool {
	return m.disabled
}

func (m *MockProvider) GetDisableReason() string {
	return m.disableReason
}

func (m *MockProvider) Geocode(ctx context.Context, address string) (*model.ProviderResult, error) {
	if m.result != nil {
		return m.result, nil
	}

	if m.success {
		return &model.ProviderResult{
			Coordinate: model.Coordinate{
				Latitude:  37.498095,
				Longitude: 127.027610,
			},
			AddressDetail: model.AddressDetail{
				RoadAddress:   "서울특별시 강남구 테헤란로 152",
				ParcelAddress: "서울특별시 강남구 역삼동 737",
				Zipcode:       "06236",
			},
			Success: true,
		}, nil
	}

	return &model.ProviderResult{
		Success: false,
		Error:   provider.ErrAddressNotFound,
	}, nil
}

// TestProvider_Interface Provider 인터페이스 구현 테스트
func TestProvider_Interface(t *testing.T) {
	// Provider 인터페이스를 구현하는지 확인
	var _ provider.GeocodingProvider = (*MockProvider)(nil)
	
	mock := &MockProvider{
		name:      "TestProvider",
		success:   true,
		available: true,
	}
	
	// 메서드 호출 테스트
	if mock.Name() != "TestProvider" {
		t.Errorf("Expected name TestProvider, got %s", mock.Name())
	}
	
	if !mock.IsAvailable(context.Background()) {
		t.Error("Expected provider to be available")
	}
	
	result, err := mock.Geocode(context.Background(), "test address")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.Success {
		t.Error("Expected successful result")
	}
}

// TestErrorTypes 에러 타입 분류 테스트
func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name      string
		errorType provider.ErrorType
		fallback  bool
		retriable bool
	}{
		{
			name:      "NotFound allows fallback",
			errorType: provider.ErrorTypeNotFound,
			fallback:  true,
			retriable: true,
		},
		{
			name:      "Invalid prevents fallback",
			errorType: provider.ErrorTypeInvalid,
			fallback:  false,
			retriable: false,
		},
		{
			name:      "SystemFailure allows fallback",
			errorType: provider.ErrorTypeSystemFailure,
			fallback:  true,
			retriable: true,
		},
		{
			name:      "Timeout allows fallback",
			errorType: provider.ErrorTypeTimeout,
			fallback:  true,
			retriable: true,
		},
		{
			name:      "RateLimitExceeded allows fallback",
			errorType: provider.ErrorTypeRateLimitExceeded,
			fallback:  true,
			retriable: true,
		},
		{
			name:      "Unauthorized prevents fallback",
			errorType: provider.ErrorTypeUnauthorized,
			fallback:  false,
			retriable: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := provider.NewClassifiedError(tt.errorType, "test error", nil)
			
			if err.Fallback != tt.fallback {
				t.Errorf("Expected fallback=%v, got %v", tt.fallback, err.Fallback)
			}
			
			if err.Retriable != tt.retriable {
				t.Errorf("Expected retriable=%v, got %v", tt.retriable, err.Retriable)
			}
		})
	}
}

// TestProviderResult_EmptyAddress 빈 주소 처리 테스트
func TestProviderResult_EmptyAddress(t *testing.T) {
	mock := &MockProvider{
		name: "TestProvider",
		result: &model.ProviderResult{
			Success: false,
			Error:   provider.ErrInvalidAddress,
		},
	}
	
	result, err := mock.Geocode(context.Background(), "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.Success {
		t.Error("Expected failure for empty address")
	}
	
	if result.Error != provider.ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress, got %v", result.Error)
	}
}

// TestProviderResult_Coordinates 좌표 검증 테스트
func TestProviderResult_Coordinates(t *testing.T) {
	result := &model.ProviderResult{
		Coordinate: model.Coordinate{
			Latitude:  37.498095,
			Longitude: 127.027610,
		},
		Success: true,
	}
	
	// 한국 좌표 범위 확인 (대략적인 범위)
	if result.Coordinate.Latitude < 33 || result.Coordinate.Latitude > 43 {
		t.Errorf("Latitude %f is outside Korea range", result.Coordinate.Latitude)
	}
	
	if result.Coordinate.Longitude < 124 || result.Coordinate.Longitude > 132 {
		t.Errorf("Longitude %f is outside Korea range", result.Coordinate.Longitude)
	}
}

// TestClassifiedError_IsClassifiedError 분류된 에러 확인 테스트
func TestClassifiedError_IsClassifiedError(t *testing.T) {
	// ClassifiedError 생성
	classifiedErr := provider.NewClassifiedError(
		provider.ErrorTypeNotFound,
		"Address not found",
		provider.ErrAddressNotFound,
	)
	
	// IsClassifiedError로 확인
	ce, ok := provider.IsClassifiedError(classifiedErr)
	if !ok {
		t.Error("Expected IsClassifiedError to return true")
	}
	
	if ce.Type != provider.ErrorTypeNotFound {
		t.Errorf("Expected ErrorTypeNotFound, got %v", ce.Type)
	}
	
	// 일반 에러는 false 반환
	normalErr := provider.ErrAddressNotFound
	_, ok = provider.IsClassifiedError(normalErr)
	if ok {
		t.Error("Expected IsClassifiedError to return false for normal error")
	}
}