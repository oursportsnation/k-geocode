package provider

import (
	"context"
	"github.com/oursportsnation/k-geocode/internal/model"
)

// GeocodingProvider 지오코딩 제공자 인터페이스
type GeocodingProvider interface {
	// Name Provider의 고유 이름 반환
	Name() string
	
	// Geocode 주소를 좌표로 변환
	// 결과가 없으면 Success=false 반환
	// 시스템 오류 발생 시 error 반환
	Geocode(ctx context.Context, address string) (*model.ProviderResult, error)
	
	// IsAvailable Provider 사용 가능 여부 확인
	// Circuit Breaker 상태 등을 체크
	IsAvailable(ctx context.Context) bool
}

// DailyLimits Provider별 일일 할당량
var DailyLimits = map[string]int{
	"vWorld": 40000,  // 일 4만건
	"Kakao":  100000, // 일 10만건
}