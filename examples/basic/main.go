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

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	geocoding "github.com/oursportsnation/k-geocode"
)

func main() {
	// 1. 설정 생성
	cfg := geocoding.DefaultConfig()
	cfg.VWorldAPIKey = os.Getenv("VWORLD_API_KEY")
	cfg.KakaoAPIKey = os.Getenv("KAKAO_API_KEY")
	cfg.Timeout = 10 * time.Second
	cfg.LogLevel = "info"

	// 2. 클라이언트 생성
	client, err := geocoding.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 3. 사용 가능한 Provider 확인
	fmt.Println("Available providers:", client.GetProviders())

	// 4. 단일 주소 지오코딩
	fmt.Println("\n=== Single Address Geocoding ===")
	result, err := client.Geocode(context.Background(), "서울특별시 중구 세종대로 110")
	if err != nil {
		log.Printf("Geocoding failed: %v", err)
	} else {
		fmt.Printf("Address: 서울특별시 중구 세종대로 110\n")
		fmt.Printf("Coordinates: (%.6f, %.6f)\n", result.Latitude, result.Longitude)
		fmt.Printf("Provider: %s\n", result.Provider)

		if result.AddressDetail != nil {
			fmt.Printf("Road Address: %s\n", result.AddressDetail.RoadAddress)
			fmt.Printf("Parcel Address: %s\n", result.AddressDetail.ParcelAddress)
		}

		// Provider 시도 내역 출력
		if len(result.Attempts) > 0 {
			fmt.Println("\nProvider Attempts:")
			for _, attempt := range result.Attempts {
				status := "✓"
				if !attempt.Success {
					status = "✗"
				}
				fmt.Printf("  %s %s\n", status, attempt.Provider)
			}
		}
	}

	// 5. 특정 주소 타입으로 지오코딩
	fmt.Println("\n=== Geocoding with Address Type ===")
	result, err = client.GeocodeWithType(
		context.Background(),
		"서울특별시 강남구 테헤란로 152",
		geocoding.AddressTypeRoad,
	)
	if err != nil {
		log.Printf("Geocoding failed: %v", err)
	} else {
		fmt.Printf("Address: 서울특별시 강남구 테헤란로 152\n")
		fmt.Printf("Coordinates: (%.6f, %.6f)\n", result.Latitude, result.Longitude)
		fmt.Printf("Provider: %s\n", result.Provider)
	}

	// 6. 배치 지오코딩
	fmt.Println("\n=== Batch Geocoding ===")
	addresses := []string{
		"서울특별시 중구 세종대로 110",
		"부산광역시 해운대구 해운대해변로 264",
		"대구광역시 중구 공평로 88",
	}

	results, err := client.GeocodeBatch(context.Background(), addresses)
	if err != nil {
		log.Fatalf("Batch geocoding failed: %v", err)
	}

	fmt.Printf("Processed %d addresses:\n", len(results))
	for i, result := range results {
		if result != nil {
			fmt.Printf("%d. %s -> (%.6f, %.6f) [%s]\n",
				i+1,
				addresses[i],
				result.Latitude,
				result.Longitude,
				result.Provider,
			)
		} else {
			fmt.Printf("%d. %s -> Failed\n", i+1, addresses[i])
		}
	}

	// 7. 서비스 가용성 확인
	fmt.Println("\n=== Service Availability ===")
	if client.IsAvailable(context.Background()) {
		fmt.Println("Service is available ✓")
	} else {
		fmt.Println("Service is unavailable ✗")
	}
}
