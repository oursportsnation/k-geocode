package utils

import (
	"fmt"
	"math"
)

// RoundToSixDecimal 소수점 6자리로 반올림 (Decimal 9,6 포맷)
// 예: 37.123456789 → 37.123457
// 예: 127.987654321 → 127.987654
func RoundToSixDecimal(val float64) float64 {
	return math.Round(val*1000000) / 1000000
}

// FormatCoordinate 좌표를 포맷팅된 문자열로 반환
func FormatCoordinate(latitude, longitude float64) string {
	return fmt.Sprintf("%.6f, %.6f", latitude, longitude)
}

// ValidateCoordinate 좌표 범위 검증 (WGS84)
// 위도: -90 ~ 90, 경도: -180 ~ 180
func ValidateCoordinate(latitude, longitude float64) bool {
	return latitude >= -90 && latitude <= 90 &&
		longitude >= -180 && longitude <= 180
}

// CalculateDistance 두 좌표 간 거리 계산 (Haversine 공식)
// 반환값: 킬로미터 (km)
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // 킬로미터
	
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	
	return earthRadius * c
}

// IsValidKoreanCoordinate 한국 영역 내 좌표인지 확인
// 한국 대략 범위: 위도 33~43, 경도 124~132
func IsValidKoreanCoordinate(latitude, longitude float64) bool {
	return latitude >= 33.0 && latitude <= 43.0 &&
		longitude >= 124.0 && longitude <= 132.0
}