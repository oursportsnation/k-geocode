package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"trim spaces", "  서울시 중구  ", "서울시 중구"},
		{"multiple spaces", "서울시    중구", "서울시 중구"},
		{"tabs and newlines", "서울시\t중구\n강남", "서울시 중구 강남"},
		{"full-width space", "서울시　중구", "서울시 중구"},
		{"already normalized", "서울시 중구", "서울시 중구"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid Seoul address", "서울특별시 중구 세종대로 110", true},
		{"valid short address", "서울시 강남구", true},
		{"valid with numbers", "서울시 중구 세종대로 110번길", true},
		{"short but valid Korean", "서울", true}, // 2 chars with Korean
		{"empty", "", false},
		{"only spaces", "   ", false},
		{"minimum valid length", "서울시 중", true},
		{"no Korean chars", "abc", false}, // no Korean characters
		{"single Korean char", "서", false}, // less than 2 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractZipcode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"5-digit zipcode", "서울시 중구 04524", "04524"},
		{"zipcode in middle", "04524 서울시 중구", "04524"},
		{"no zipcode", "서울시 중구 세종대로", ""},
		{"6-digit (not valid)", "서울시 중구 045244", ""},
		{"multiple numbers", "서울시 110 중구 04524", "04524"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractZipcode(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSplitAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"basic split", "서울시 중구 세종대로", []string{"서울시", "중구", "세종대로"}},
		{"single part", "서울시", []string{"서울시"}},
		{"empty", "", nil}, // strings.Fields returns nil for empty string
		{"multiple spaces", "서울시  중구", []string{"서울시", "중구"}}, // Fields ignores consecutive spaces
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
