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

// AddressType represents the type of Korean address format.
type AddressType string

const (
	// AddressTypeRoad represents a road-based address (도로명 주소).
	AddressTypeRoad AddressType = "ROAD"

	// AddressTypeParcel represents a parcel-based address (지번 주소).
	AddressTypeParcel AddressType = "PARCEL"
)

// Result represents a geocoding result containing WGS84 coordinates.
type Result struct {
	// Latitude is the WGS84 latitude coordinate.
	Latitude float64 `json:"latitude"`

	// Longitude is the WGS84 longitude coordinate.
	Longitude float64 `json:"longitude"`

	// Provider is the name of the provider that returned this result (e.g., "vWorld", "Kakao").
	Provider string `json:"provider"`

	// AddressDetail contains additional address information if available.
	AddressDetail *AddressDetail `json:"address_detail,omitempty"`

	// Attempts contains the list of provider attempts made during geocoding.
	Attempts []Attempt `json:"attempts,omitempty"`
}

// AddressDetail contains detailed address information returned by the provider.
type AddressDetail struct {
	// RoadAddress is the road-based address (도로명 주소).
	RoadAddress string `json:"road_address,omitempty"`

	// ParcelAddress is the parcel-based address (지번 주소).
	ParcelAddress string `json:"parcel_address,omitempty"`

	// BuildingName is the name of the building, if applicable.
	BuildingName string `json:"building_name,omitempty"`

	// Zipcode is the postal code.
	Zipcode string `json:"zipcode,omitempty"`
}

// Attempt records a single provider attempt during the geocoding process.
type Attempt struct {
	// Provider is the name of the provider that was tried.
	Provider string `json:"provider"`

	// Success indicates whether this attempt succeeded.
	Success bool `json:"success"`

	// Error contains the error message if the attempt failed.
	Error string `json:"error,omitempty"`
}
