package unit

import (
	"math"
	"testing"
	
	"github.com/oursportsnation/k-geocode/internal/utils"
)

// TestRoundToSixDecimal 소수점 6자리 반올림 테스트
func TestRoundToSixDecimal(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{37.123456789, 37.123457},
		{127.987654321, 127.987654},
		{37.1234564, 37.123456},   // 반올림 내림
		{37.1234565, 37.123457},   // 반올림 올림
		{37.123456, 37.123456},    // 정확히 6자리
		{37.1, 37.1},              // 소수점 1자리
		{37, 37},                  // 정수
		{-37.123456789, -37.123457}, // 음수
	}
	
	for _, tt := range tests {
		result := utils.RoundToSixDecimal(tt.input)
		if result != tt.expected {
			t.Errorf("RoundToSixDecimal(%f) = %f; want %f", tt.input, result, tt.expected)
		}
	}
}

// TestValidateCoordinate 좌표 유효성 검증 테스트
func TestValidateCoordinate(t *testing.T) {
	tests := []struct {
		lat   float64
		lon   float64
		valid bool
		desc  string
	}{
		{37.5, 127.0, true, "valid Seoul coordinates"},
		{90, 180, true, "max valid coordinates"},
		{-90, -180, true, "min valid coordinates"},
		{91, 127, false, "latitude too high"},
		{-91, 127, false, "latitude too low"},
		{37, 181, false, "longitude too high"},
		{37, -181, false, "longitude too low"},
		{0, 0, true, "zero coordinates"},
	}
	
	for _, tt := range tests {
		result := utils.ValidateCoordinate(tt.lat, tt.lon)
		if result != tt.valid {
			t.Errorf("ValidateCoordinate(%f, %f) = %v; want %v (%s)",
				tt.lat, tt.lon, result, tt.valid, tt.desc)
		}
	}
}

// TestIsValidKoreanCoordinate 한국 영역 좌표 테스트
func TestIsValidKoreanCoordinate(t *testing.T) {
	tests := []struct {
		lat   float64
		lon   float64
		valid bool
		desc  string
	}{
		{37.566535, 126.977969, true, "Seoul"},
		{35.179554, 129.075641, true, "Busan"},
		{33.499621, 126.531188, true, "Jeju"},
		{38.0, 127.5, true, "DMZ area"},
		{32.0, 127.0, false, "too south"},
		{44.0, 127.0, false, "too north"},
		{37.0, 123.0, false, "too west"},
		{37.0, 133.0, false, "too east"},
		{0, 0, false, "Africa"},
		{40.7128, -74.0060, false, "New York"},
	}
	
	for _, tt := range tests {
		result := utils.IsValidKoreanCoordinate(tt.lat, tt.lon)
		if result != tt.valid {
			t.Errorf("IsValidKoreanCoordinate(%f, %f) = %v; want %v (%s)",
				tt.lat, tt.lon, result, tt.valid, tt.desc)
		}
	}
}

// TestCalculateDistance 거리 계산 테스트
func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		lat1     float64
		lon1     float64
		lat2     float64
		lon2     float64
		expected float64 // km
		delta    float64 // 허용 오차
		desc     string
	}{
		{37.566535, 126.977969, 37.566535, 126.977969, 0, 0.01, "same point"},
		{37.566535, 126.977969, 35.179554, 129.075641, 325, 5, "Seoul to Busan"},
		{37.566535, 126.977969, 33.499621, 126.531188, 452, 5, "Seoul to Jeju"},
		{0, 0, 0, 1, 111.32, 1, "1 degree longitude at equator"},
		{90, 0, -90, 0, 20015.0, 100, "pole to pole"},
	}
	
	for _, tt := range tests {
		result := utils.CalculateDistance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
		diff := math.Abs(result - tt.expected)
		if diff > tt.delta {
			t.Errorf("CalculateDistance(%f,%f,%f,%f) = %f; want %f±%f (%s)",
				tt.lat1, tt.lon1, tt.lat2, tt.lon2, result, tt.expected, tt.delta, tt.desc)
		}
	}
}

// TestNormalizeAddress 주소 정규화 테스트
func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  서울시   강남구  ", "서울시 강남구"},
		{"서울시\t\n강남구", "서울시 강남구"},
		{"서울시　　강남구", "서울시 강남구"}, // 전각 공백
		{"서울시（강남구）", "서울시(강남구)"},  // 전각 괄호
		{"서울시，강남구", "서울시,강남구"},     // 전각 쉼표
		{"", ""},
		{"   ", ""},
	}
	
	for _, tt := range tests {
		result := utils.NormalizeAddress(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeAddress(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

// TestIsValidAddress 주소 유효성 검증 테스트
func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		desc  string
	}{
		{"서울시 강남구", true, "valid Korean address"},
		{"경기도 성남시 분당구 판교동", true, "valid long address"},
		{"제주특별자치도", true, "valid region name"},
		{"Seoul Gangnam", false, "English only"},
		{"123456", false, "numbers only"},
		{"", false, "empty string"},
		{"   ", false, "whitespace only"},
		{"!", false, "single char"},
		{"서", false, "single Korean char"},
		{"서울 123번지", true, "Korean with numbers"},
		{"@#$%", false, "special chars only"},
	}
	
	for _, tt := range tests {
		result := utils.IsValidAddress(tt.input)
		if result != tt.valid {
			t.Errorf("IsValidAddress(%q) = %v; want %v (%s)",
				tt.input, result, tt.valid, tt.desc)
		}
	}
}

// TestExtractZipcode 우편번호 추출 테스트
func TestExtractZipcode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"서울시 강남구 06236", "06236"},
		{"06236 서울시 강남구", "06236"},
		{"서울시 06236 강남구", "06236"},
		{"서울시 강남구 테헤란로 152 (06236)", "06236"},
		{"서울시 강남구", ""},
		{"12345678", ""},    // 8자리는 우편번호가 아님
		{"1234", ""},        // 4자리는 우편번호가 아님
		{"우편번호 06236", "06236"},
	}
	
	for _, tt := range tests {
		result := utils.ExtractZipcode(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractZipcode(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

// TestFormatCoordinate 좌표 포맷팅 테스트
func TestFormatCoordinate(t *testing.T) {
	tests := []struct {
		lat      float64
		lon      float64
		expected string
	}{
		{37.566535, 126.977969, "37.566535, 126.977969"},
		{35.179554, 129.075641, "35.179554, 129.075641"},
		{37.5, 127.0, "37.500000, 127.000000"},
		{-12.345678, -98.765432, "-12.345678, -98.765432"},
		{0, 0, "0.000000, 0.000000"},
	}
	
	for _, tt := range tests {
		result := utils.FormatCoordinate(tt.lat, tt.lon)
		if result != tt.expected {
			t.Errorf("FormatCoordinate(%f, %f) = %q; want %q",
				tt.lat, tt.lon, result, tt.expected)
		}
	}
}