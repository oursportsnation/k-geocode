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

// Package geocoding provides a unified Korean geocoding client that supports
// multiple providers (vWorld, Kakao) with automatic fallback.
//
// # Features
//
//   - Multi-provider support: vWorld and Kakao geocoding APIs
//   - Automatic fallback: Tries next provider on failure
//   - Address type handling: Supports both road (도로명) and parcel (지번) addresses
//   - Batch processing: Process up to 100 addresses concurrently
//   - WGS84 coordinates: Returns standard GPS coordinates
//
// # Quick Start
//
// Create a client with your API keys:
//
//	cfg := geocoding.DefaultConfig()
//	cfg.VWorldAPIKey = "your-vworld-key"
//	cfg.KakaoAPIKey = "your-kakao-key"
//
//	client, err := geocoding.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
// Geocode a single address:
//
//	result, err := client.Geocode(ctx, "서울특별시 중구 세종대로 110")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Coordinates: %.6f, %.6f\n", result.Latitude, result.Longitude)
//
// Geocode multiple addresses:
//
//	addresses := []string{
//	    "서울특별시 중구 세종대로 110",
//	    "부산광역시 해운대구 해운대해변로 264",
//	}
//	results, err := client.GeocodeBatch(ctx, addresses)
//
// # Provider Priority
//
// By default, providers are tried in the following order:
//  1. vWorld (if API key provided)
//  2. Kakao (if API key provided)
//
// If a provider fails (network error, invalid response, etc.),
// the next provider is automatically tried.
//
// # Address Types
//
// Korean addresses can be specified as road addresses (도로명) or
// parcel addresses (지번). Use [Client.GeocodeWithType] to specify:
//
//	result, err := client.GeocodeWithType(ctx, address, geocoding.AddressTypeRoad)
//
// If no type is specified, both types are tried automatically.
package geocoding
