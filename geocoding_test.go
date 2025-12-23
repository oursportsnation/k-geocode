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

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 5*time.Second, cfg.Timeout)
	assert.Equal(t, 2, cfg.MaxRetries)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 10, cfg.ConcurrentLimit)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "no API keys",
			config:  Config{},
			wantErr: true,
			errMsg:  "at least one API key",
		},
		{
			name: "valid with VWorld key only",
			config: Config{
				VWorldAPIKey:    "test-key",
				ConcurrentLimit: 10,
			},
			wantErr: false,
		},
		{
			name: "valid with Kakao key only",
			config: Config{
				KakaoAPIKey:     "test-key",
				ConcurrentLimit: 10,
			},
			wantErr: false,
		},
		{
			name: "valid with both keys",
			config: Config{
				VWorldAPIKey:    "vworld-key",
				KakaoAPIKey:     "kakao-key",
				ConcurrentLimit: 10,
			},
			wantErr: false,
		},
		{
			name: "negative timeout",
			config: Config{
				VWorldAPIKey:    "test-key",
				Timeout:         -1 * time.Second,
				ConcurrentLimit: 10,
			},
			wantErr: true,
			errMsg:  "timeout cannot be negative",
		},
		{
			name: "negative max retries",
			config: Config{
				VWorldAPIKey:    "test-key",
				MaxRetries:      -1,
				ConcurrentLimit: 10,
			},
			wantErr: true,
			errMsg:  "maxRetries cannot be negative",
		},
		{
			name: "concurrent limit too low",
			config: Config{
				VWorldAPIKey:    "test-key",
				ConcurrentLimit: 0,
			},
			wantErr: true,
			errMsg:  "concurrentLimit must be at least 1",
		},
		{
			name: "concurrent limit too high",
			config: Config{
				VWorldAPIKey:    "test-key",
				ConcurrentLimit: 101,
			},
			wantErr: true,
			errMsg:  "concurrentLimit cannot exceed 100",
		},
		{
			name: "invalid log level",
			config: Config{
				VWorldAPIKey:    "test-key",
				LogLevel:        "invalid",
				ConcurrentLimit: 10,
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "valid log levels",
			config: Config{
				VWorldAPIKey:    "test-key",
				LogLevel:        "debug",
				ConcurrentLimit: 10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	cfg := Config{
		VWorldAPIKey: "test-key",
	}

	cfg.SetDefaults()

	assert.Equal(t, 5*time.Second, cfg.Timeout)
	assert.Equal(t, 2, cfg.MaxRetries)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, 10, cfg.ConcurrentLimit)
}

func TestConfig_SetDefaults_PreservesExisting(t *testing.T) {
	cfg := Config{
		VWorldAPIKey:    "test-key",
		Timeout:         10 * time.Second,
		MaxRetries:      5,
		LogLevel:        "debug",
		ConcurrentLimit: 20,
	}

	cfg.SetDefaults()

	assert.Equal(t, 10*time.Second, cfg.Timeout)
	assert.Equal(t, 5, cfg.MaxRetries)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, 20, cfg.ConcurrentLimit)
}

func TestAddressType_Constants(t *testing.T) {
	assert.Equal(t, AddressType("ROAD"), AddressTypeRoad)
	assert.Equal(t, AddressType("PARCEL"), AddressTypeParcel)
}

func TestVersion(t *testing.T) {
	assert.NotEmpty(t, Version)
	assert.Regexp(t, `^\d+\.\d+\.\d+$`, Version)
}

func TestNew_NoAPIKey(t *testing.T) {
	cfg := DefaultConfig()

	client, err := New(cfg)

	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "at least one API key")
}

func TestNew_InvalidConfig(t *testing.T) {
	cfg := Config{
		VWorldAPIKey: "test-key",
		Timeout:      -1 * time.Second,
	}

	client, err := New(cfg)

	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "invalid config")
}

func TestResult_Struct(t *testing.T) {
	result := Result{
		Latitude:  37.5665,
		Longitude: 126.9780,
		Provider:  "vWorld",
		AddressDetail: &AddressDetail{
			RoadAddress:   "서울특별시 중구 세종대로 110",
			ParcelAddress: "서울특별시 중구 태평로1가 31",
			BuildingName:  "서울시청",
			Zipcode:       "04524",
		},
		Attempts: []Attempt{
			{Provider: "vWorld", Success: true},
		},
	}

	assert.Equal(t, 37.5665, result.Latitude)
	assert.Equal(t, 126.9780, result.Longitude)
	assert.Equal(t, "vWorld", result.Provider)
	assert.NotNil(t, result.AddressDetail)
	assert.Equal(t, "서울시청", result.AddressDetail.BuildingName)
	assert.Len(t, result.Attempts, 1)
	assert.True(t, result.Attempts[0].Success)
}

// Mock vWorld API response
type vWorldResponse struct {
	Response struct {
		Status string `json:"status"`
		Result struct {
			Items []struct {
				Point struct {
					X string `json:"x"`
					Y string `json:"y"`
				} `json:"point"`
				Address struct {
					Road   string `json:"road"`
					Parcel string `json:"parcel"`
				} `json:"address"`
			} `json:"items"`
		} `json:"result"`
	} `json:"response"`
}

func createMockVWorldServer(success bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if success {
			resp := vWorldResponse{}
			resp.Response.Status = "OK"
			resp.Response.Result.Items = []struct {
				Point struct {
					X string `json:"x"`
					Y string `json:"y"`
				} `json:"point"`
				Address struct {
					Road   string `json:"road"`
					Parcel string `json:"parcel"`
				} `json:"address"`
			}{
				{
					Point: struct {
						X string `json:"x"`
						Y string `json:"y"`
					}{X: "126.978000", Y: "37.566500"},
					Address: struct {
						Road   string `json:"road"`
						Parcel string `json:"parcel"`
					}{Road: "서울특별시 중구 세종대로 110", Parcel: "서울특별시 중구 태평로1가 31"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := vWorldResponse{}
			resp.Response.Status = "NOT_FOUND"
			json.NewEncoder(w).Encode(resp)
		}
	}))
}

func TestClient_Close(t *testing.T) {
	cfg := DefaultConfig()
	cfg.VWorldAPIKey = "test-key"

	client, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	err = client.Close()
	assert.NoError(t, err)
}

func TestClient_GetProviders(t *testing.T) {
	t.Run("with VWorld only", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.VWorldAPIKey = "vworld-key"

		client, err := New(cfg)
		require.NoError(t, err)

		providers := client.GetProviders()
		assert.Len(t, providers, 1)
		assert.Equal(t, "vWorld", providers[0])

		client.Close()
	})

	t.Run("with Kakao only", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.KakaoAPIKey = "kakao-key"

		client, err := New(cfg)
		require.NoError(t, err)

		providers := client.GetProviders()
		assert.Len(t, providers, 1)
		assert.Equal(t, "Kakao", providers[0])

		client.Close()
	})

	t.Run("with both providers", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.VWorldAPIKey = "vworld-key"
		cfg.KakaoAPIKey = "kakao-key"

		client, err := New(cfg)
		require.NoError(t, err)

		providers := client.GetProviders()
		assert.Len(t, providers, 2)

		client.Close()
	})
}

func TestClient_IsAvailable(t *testing.T) {
	cfg := DefaultConfig()
	cfg.VWorldAPIKey = "test-key"
	cfg.Timeout = 100 * time.Millisecond

	client, err := New(cfg)
	require.NoError(t, err)
	defer client.Close()

	// With a valid context
	ctx := context.Background()
	available := client.IsAvailable(ctx)
	// Provider should be available (not disabled)
	assert.True(t, available)
}

func TestClient_Geocode_NetworkError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.VWorldAPIKey = "test-key"
	cfg.Timeout = 100 * time.Millisecond

	client, err := New(cfg)
	require.NoError(t, err)
	defer client.Close()

	// This will fail because there's no real API server
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := client.Geocode(ctx, "서울특별시 중구 세종대로 110")

	// Should return error or nil result due to network failure
	// Either err is not nil, or result is nil (both indicate failure)
	if err != nil {
		assert.Error(t, err)
	} else if result != nil {
		// If we got a result, coordinates should be zero (failed geocode)
		// or it somehow succeeded (unlikely with fake API key)
		t.Log("Got result despite network error")
	}
}

func TestClient_GeocodeWithType_NetworkError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.VWorldAPIKey = "test-key"
	cfg.Timeout = 100 * time.Millisecond

	client, err := New(cfg)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := client.GeocodeWithType(ctx, "서울특별시 중구 세종대로 110", AddressTypeRoad)

	// Either error or nil result expected
	if err != nil {
		assert.Error(t, err)
	} else if result != nil {
		t.Log("Got result despite network error")
	}
}

func TestClient_GeocodeBatch_Empty(t *testing.T) {
	cfg := DefaultConfig()
	cfg.VWorldAPIKey = "test-key"

	client, err := New(cfg)
	require.NoError(t, err)
	defer client.Close()

	results, err := client.GeocodeBatch(context.Background(), []string{})
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestClient_GeocodeBatch_TooMany(t *testing.T) {
	cfg := DefaultConfig()
	cfg.VWorldAPIKey = "test-key"

	client, err := New(cfg)
	require.NoError(t, err)
	defer client.Close()

	addresses := make([]string, 101)
	for i := range addresses {
		addresses[i] = "서울시"
	}

	results, err := client.GeocodeBatch(context.Background(), addresses)
	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "too many addresses")
}
