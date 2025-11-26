package unit

import (
	"context"
	"testing"
	"time"
	
	"github.com/oursportsnation/k-geocode/internal/model"
	"github.com/oursportsnation/k-geocode/internal/provider"
	"github.com/oursportsnation/k-geocode/internal/service"
	
	"go.uber.org/zap"
)

// Import fmt for Sprintf
import "fmt"

// TestGeocoding_Fallback 폴백 로직 테스트
func TestGeocoding_Fallback(t *testing.T) {
	logger := zap.NewNop()
	
	providers := []provider.GeocodingProvider{
		&MockProvider{name: "Provider1", success: false, available: true},
		&MockProvider{name: "Provider2", success: true, available: true},
	}
	
	svc := service.NewGeocodingService(providers, logger)
	
	resp, err := svc.Geocode(context.Background(), "서울시 강남구", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if !resp.Success {
		t.Error("Expected success")
	}
	
	if resp.Provider != "Provider2" {
		t.Errorf("Expected Provider2, got %s", resp.Provider)
	}
}

// TestGeocoding_CoordinatePrecision 좌표 정밀도 테스트
func TestGeocoding_CoordinatePrecision(t *testing.T) {
	logger := zap.NewNop()
	
	// 소수점 많은 좌표를 반환하는 Mock Provider
	mockResult := &model.ProviderResult{
		Coordinate: model.Coordinate{
			Latitude:  37.123456789,
			Longitude: 127.987654321,
		},
		Success: true,
	}
	
	providers := []provider.GeocodingProvider{
		&MockProvider{
			name:      "TestProvider",
			success:   true,
			available: true,
			result:    mockResult,
		},
	}
	
	svc := service.NewGeocodingService(providers, logger)
	
	resp, err := svc.Geocode(context.Background(), "서울시 강남구", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !resp.Success || resp.Coordinate == nil {
		t.Fatal("Expected successful response with coordinates")
	}
	
	// 소수점 6자리 검증
	lat := resp.Coordinate.Latitude
	expected := 37.123457 // 37.123456789 → 37.123457 (반올림)
	
	if lat != expected {
		t.Errorf("Expected %f, got %f", expected, lat)
	}
	
	lng := resp.Coordinate.Longitude
	expectedLng := 127.987654 // 127.987654321 → 127.987654
	
	if lng != expectedLng {
		t.Errorf("Expected longitude %f, got %f", expectedLng, lng)
	}
}

// TestGeocoding_EmptyAddress 빈 주소 처리 테스트
func TestGeocoding_EmptyAddress(t *testing.T) {
	logger := zap.NewNop()
	svc := service.NewGeocodingService(nil, logger)
	
	testCases := []string{
		"",
		"   ",
		"\t\n",
	}
	
	for _, tc := range testCases {
		resp, err := svc.Geocode(context.Background(), tc, "")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if resp.Success {
			t.Errorf("Expected failure for empty address: %q", tc)
		}

		if resp.Error == "" {
			t.Error("Expected error message")
		}
	}
}

// TestGeocoding_InvalidAddress 잘못된 주소 테스트
func TestGeocoding_InvalidAddress(t *testing.T) {
	logger := zap.NewNop()
	svc := service.NewGeocodingService(nil, logger)
	
	testCases := []string{
		"123",        // 숫자만
		"abc",        // 영문만
		"!@#$%",      // 특수문자만
		"a",          // 너무 짧음
	}
	
	for _, tc := range testCases {
		resp, _ := svc.Geocode(context.Background(), tc, "")

		if resp.Success {
			t.Errorf("Expected failure for invalid address: %q", tc)
		}
	}
}

// TestGeocoding_AllProvidersFailed 모든 Provider 실패 테스트
func TestGeocoding_AllProvidersFailed(t *testing.T) {
	logger := zap.NewNop()
	
	providers := []provider.GeocodingProvider{
		&MockProvider{name: "Provider1", success: false, available: true},
		&MockProvider{name: "Provider2", success: false, available: true},
	}
	
	svc := service.NewGeocodingService(providers, logger)
	
	resp, err := svc.Geocode(context.Background(), "서울시 강남구", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if resp.Success {
		t.Error("Expected failure when all providers fail")
	}
	
	if resp.Provider != "none" {
		t.Errorf("Expected provider 'none', got %s", resp.Provider)
	}
}

// TestGeocoding_ProviderNotAvailable Provider 사용 불가 테스트
func TestGeocoding_ProviderNotAvailable(t *testing.T) {
	logger := zap.NewNop()
	
	providers := []provider.GeocodingProvider{
		&MockProvider{name: "Provider1", success: true, available: false}, // 사용 불가
		&MockProvider{name: "Provider2", success: true, available: true},  // 사용 가능
	}
	
	svc := service.NewGeocodingService(providers, logger)
	
	resp, err := svc.Geocode(context.Background(), "서울시 강남구", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !resp.Success {
		t.Error("Expected success")
	}
	
	// 사용 가능한 Provider2가 사용되어야 함
	if resp.Provider != "Provider2" {
		t.Errorf("Expected Provider2, got %s", resp.Provider)
	}
}

// TestGeocodeBatch_Success 배치 처리 성공 테스트
func TestGeocodeBatch_Success(t *testing.T) {
	logger := zap.NewNop()
	
	providers := []provider.GeocodingProvider{
		&MockProvider{name: "TestProvider", success: true, available: true},
	}
	
	svc := service.NewGeocodingService(providers, logger)
	
	addresses := []string{
		"서울시 강남구 테헤란로",
		"경기도 성남시 분당구",
		"부산시 해운대구",
	}
	
	resp, err := svc.GeocodeBatch(context.Background(), addresses)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if resp.Summary.Total != 3 {
		t.Errorf("Expected total 3, got %d", resp.Summary.Total)
	}
	
	if resp.Summary.Success != 3 {
		t.Errorf("Expected success 3, got %d", resp.Summary.Success)
	}
	
	if resp.Summary.Failed != 0 {
		t.Errorf("Expected failed 0, got %d", resp.Summary.Failed)
	}
	
	// 결과 순서 확인
	for i, result := range resp.Results {
		if !result.Success {
			t.Errorf("Expected result %d to be successful", i)
		}
	}
}

// TestGeocodeBatch_PartialFailure 배치 처리 일부 실패 테스트
func TestGeocodeBatch_PartialFailure(t *testing.T) {
	logger := zap.NewNop()
	
	// 일부만 성공하는 Provider
	providers := []provider.GeocodingProvider{
		&MockProvider{
			name:      "TestProvider",
			success:   true,
			available: true,
			result: &model.ProviderResult{
				Success: false, // 기본적으로 실패
			},
		},
	}
	
	svc := service.NewGeocodingService(providers, logger)
	
	addresses := []string{
		"",                // 빈 주소 (실패)
		"서울시 강남구",   // 정상 주소
		"!@#$%",          // 잘못된 주소 (실패)
	}
	
	resp, err := svc.GeocodeBatch(context.Background(), addresses)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if resp.Summary.Total != 3 {
		t.Errorf("Expected total 3, got %d", resp.Summary.Total)
	}
	
	// 첫 번째와 세 번째는 실패해야 함
	if resp.Results[0].Success {
		t.Error("Expected first result to fail")
	}
	
	if resp.Results[2].Success {
		t.Error("Expected third result to fail")
	}
}

// TestGeocodeBatch_Empty 빈 배치 테스트
func TestGeocodeBatch_Empty(t *testing.T) {
	logger := zap.NewNop()
	svc := service.NewGeocodingService(nil, logger)
	
	resp, err := svc.GeocodeBatch(context.Background(), []string{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(resp.Results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(resp.Results))
	}
	
	if resp.Summary.Total != 0 {
		t.Errorf("Expected total 0, got %d", resp.Summary.Total)
	}
}

// TestGeocodeBatch_ConcurrentProcessing 동시 처리 테스트
func TestGeocodeBatch_ConcurrentProcessing(t *testing.T) {
	logger := zap.NewNop()
	
	// 처리 시간이 걸리는 Mock Provider
	providers := []provider.GeocodingProvider{
		&DelayedMockProvider{
			MockProvider: MockProvider{
				name:      "SlowProvider",
				success:   true,
				available: true,
			},
			delay: 10 * time.Millisecond,
		},
	}
	
	svc := service.NewGeocodingService(providers, logger)
	
	// 많은 수의 주소
	addresses := make([]string, 20)
	for i := range addresses {
		addresses[i] = fmt.Sprintf("주소 %d", i)
	}
	
	start := time.Now()
	resp, err := svc.GeocodeBatch(context.Background(), addresses)
	elapsed := time.Since(start)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if resp.Summary.Total != 20 {
		t.Errorf("Expected total 20, got %d", resp.Summary.Total)
	}
	
	// 동시 처리로 인해 순차 처리보다 빨라야 함
	sequentialTime := 20 * 10 * time.Millisecond
	if elapsed > sequentialTime {
		t.Errorf("Concurrent processing too slow: %v", elapsed)
	}
	
	t.Logf("Batch processed in %v (sequential would be %v)", elapsed, sequentialTime)
}

// DelayedMockProvider 지연이 있는 Mock Provider (동시성 테스트용)
type DelayedMockProvider struct {
	MockProvider
	delay time.Duration
}

func (d *DelayedMockProvider) Geocode(ctx context.Context, address string) (*model.ProviderResult, error) {
	time.Sleep(d.delay)
	return d.MockProvider.Geocode(ctx, address)
}