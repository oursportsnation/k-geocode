package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoundToSixDecimal(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"positive with many decimals", 37.12345678, 37.123457},
		{"negative value", -127.12345678, -127.123457},
		{"already six decimals", 37.123456, 37.123456},
		{"fewer decimals", 37.12, 37.12},
		{"zero", 0.0, 0.0},
		{"round down", 37.1234564, 37.123456},
		{"round up", 37.1234565, 37.123457},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToSixDecimal(tt.input)
			assert.InDelta(t, tt.expected, result, 0.0000001)
		})
	}
}

func TestFormatCoordinate(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lng      float64
		expected string
	}{
		{"Seoul coordinates", 37.5665, 126.978, "37.566500, 126.978000"},
		{"negative coordinates", -33.8688, 151.2093, "-33.868800, 151.209300"},
		{"zero coordinates", 0.0, 0.0, "0.000000, 0.000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCoordinate(tt.lat, tt.lng)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateCoordinate(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lng      float64
		expected bool
	}{
		{"valid Seoul", 37.5665, 126.978, true},
		{"valid negative lat", -33.8688, 151.2093, true},
		{"invalid lat too high", 91.0, 126.978, false},
		{"invalid lat too low", -91.0, 126.978, false},
		{"invalid lng too high", 37.5665, 181.0, false},
		{"invalid lng too low", 37.5665, -181.0, false},
		{"boundary lat max", 90.0, 0.0, true},
		{"boundary lat min", -90.0, 0.0, true},
		{"boundary lng max", 0.0, 180.0, true},
		{"boundary lng min", 0.0, -180.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCoordinate(tt.lat, tt.lng)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidKoreanCoordinate(t *testing.T) {
	tests := []struct {
		name     string
		lat      float64
		lng      float64
		expected bool
	}{
		{"Seoul", 37.5665, 126.978, true},
		{"Busan", 35.1796, 129.0756, true},
		{"Jeju", 33.4996, 126.5312, true},
		{"Dokdo", 37.2426, 131.8597, true},
		{"outside Korea - Tokyo", 35.6762, 139.6503, false},
		{"outside Korea - Beijing", 39.9042, 116.4074, false},
		{"outside Korea - too far north", 44.0, 127.0, false},
		{"outside Korea - too far south", 32.0, 127.0, false},
		{"outside Korea - too far west", 37.0, 123.0, false},
		{"outside Korea - too far east", 37.0, 133.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidKoreanCoordinate(tt.lat, tt.lng)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name        string
		lat1, lng1  float64
		lat2, lng2  float64
		minExpected float64
		maxExpected float64
	}{
		{
			name:        "Seoul to Busan",
			lat1:        37.5665, lng1: 126.978,
			lat2:        35.1796, lng2: 129.0756,
			minExpected: 320.0, maxExpected: 330.0,
		},
		{
			name:        "same point",
			lat1:        37.5665, lng1: 126.978,
			lat2:        37.5665, lng2: 126.978,
			minExpected: 0.0, maxExpected: 0.001,
		},
		{
			name:        "Seoul to Incheon",
			lat1:        37.5665, lng1: 126.978,
			lat2:        37.4563, lng2: 126.7052,
			minExpected: 25.0, maxExpected: 30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			assert.GreaterOrEqual(t, result, tt.minExpected)
			assert.LessOrEqual(t, result, tt.maxExpected)
		})
	}
}
