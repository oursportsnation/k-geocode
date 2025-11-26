// Copyright 2025 Our Sports Nation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package geocoding

// AddressType 주소 타입
type AddressType string

const (
	// AddressTypeRoad 도로명 주소
	AddressTypeRoad AddressType = "ROAD"

	// AddressTypeParcel 지번 주소
	AddressTypeParcel AddressType = "PARCEL"
)

// Result 지오코딩 결과
type Result struct {
	// Latitude 위도 (WGS84)
	Latitude float64 `json:"latitude"`

	// Longitude 경도 (WGS84)
	Longitude float64 `json:"longitude"`

	// Provider 사용된 Provider (vWorld, Kakao)
	Provider string `json:"provider"`

	// AddressDetail 주소 상세 정보 (선택적)
	AddressDetail *AddressDetail `json:"address_detail,omitempty"`

	// Attempts Provider 시도 내역
	Attempts []Attempt `json:"attempts,omitempty"`
}

// AddressDetail 주소 상세 정보
type AddressDetail struct {
	// RoadAddress 도로명 주소
	RoadAddress string `json:"road_address,omitempty"`

	// ParcelAddress 지번 주소
	ParcelAddress string `json:"parcel_address,omitempty"`

	// BuildingName 건물명
	BuildingName string `json:"building_name,omitempty"`

	// Zipcode 우편번호
	Zipcode string `json:"zipcode,omitempty"`
}

// Attempt Provider 시도 정보
type Attempt struct {
	// Provider Provider 이름
	Provider string `json:"provider"`

	// Success 성공 여부
	Success bool `json:"success"`

	// Error 에러 메시지 (실패 시)
	Error string `json:"error,omitempty"`
}
