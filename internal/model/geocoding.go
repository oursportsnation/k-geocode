package model

import "time"

// GeocodingRequest 지오코딩 요청
type GeocodingRequest struct {
	Address     string `json:"address" binding:"required"`                  // 검색 주소
	AddressType string `json:"address_type,omitempty" binding:"omitempty,oneof=ROAD PARCEL road parcel"` // 주소 타입 (ROAD, PARCEL) - 선택적
}

// Coordinate 좌표 정보 (소수점 6자리 정밀도)
type Coordinate struct {
	Latitude  float64 `json:"latitude"`  // 위도 (y) - Decimal(9,6)
	Longitude float64 `json:"longitude"` // 경도 (x) - Decimal(9,6)
}

// AddressDetail 상세 주소 정보
type AddressDetail struct {
	RoadAddress   string `json:"road_address"`   // 도로명 주소
	ParcelAddress string `json:"parcel_address"` // 지번 주소
	Zipcode       string `json:"zipcode"`        // 우편번호
	BuildingName  string `json:"building_name"`  // 건물명
}

// ProviderAttempt Provider 시도 정보
type ProviderAttempt struct {
	Provider string `json:"provider"`           // Provider 이름
	Success  bool   `json:"success"`            // 성공 여부
	Error    string `json:"error,omitempty"`    // 에러 메시지
}

// GeocodingResponse 지오코딩 응답
type GeocodingResponse struct {
	Success        bool               `json:"success"`
	Coordinate     *Coordinate        `json:"coordinate,omitempty"`
	AddressDetail  *AddressDetail     `json:"address_detail,omitempty"`
	Provider       string             `json:"provider"`                                  // 최종 사용된 제공자
	Attempts       []ProviderAttempt  `json:"attempts,omitempty"`                        // Provider 시도 내역
	ProcessedAt    time.Time          `json:"processed_at"`
	ProcessingTime time.Duration      `json:"processing_time_ms" swaggertype:"integer"` // 밀리초
	Error          string             `json:"error,omitempty"`
}

// BulkRequest 대량 변환 요청
type BulkRequest struct {
	Addresses []string `json:"addresses" binding:"required,max=100"` // 최대 100건
}

// BulkResponse 대량 변환 응답
type BulkResponse struct {
	Results []*GeocodingResponse `json:"results"`
	Summary struct {
		Total   int `json:"total"`
		Success int `json:"success"`
		Failed  int `json:"failed"`
	} `json:"summary"`
	ProcessingTime time.Duration `json:"processing_time_ms" swaggertype:"integer"`
}

// ProviderResult Provider에서 반환하는 내부 결과
type ProviderResult struct {
	Coordinate    Coordinate
	AddressDetail AddressDetail
	Success       bool
	Error         error
}