package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// NormalizeAddress 주소 정규화
func NormalizeAddress(address string) string {
	// 특수문자 정규화 (전각 공백 포함)
	address = normalizeSpecialChars(address)
	
	// 공백 정리
	address = strings.TrimSpace(address)
	
	// 연속된 공백을 하나로
	space := regexp.MustCompile(`\s+`)
	address = space.ReplaceAllString(address, " ")
	
	return address
}

// normalizeSpecialChars 특수문자 정규화
func normalizeSpecialChars(s string) string {
	// 전각 문자를 반각으로
	replacer := strings.NewReplacer(
		"（", "(",
		"）", ")",
		"－", "-",
		"～", "~",
		"，", ",",
		"．", ".",
		"：", ":",
		"；", ";",
		"　", " ", // 전각 공백을 반각 공백으로
	)
	return replacer.Replace(s)
}

// IsValidAddress 주소 유효성 검증
func IsValidAddress(address string) bool {
	// 빈 문자열 체크
	if strings.TrimSpace(address) == "" {
		return false
	}
	
	// 최소 길이 체크 (최소 2자 이상)
	if len([]rune(address)) < 2 {
		return false
	}
	
	// 한글이 포함되어 있는지 체크
	hasKorean := false
	for _, r := range address {
		if unicode.Is(unicode.Hangul, r) {
			hasKorean = true
			break
		}
	}
	
	return hasKorean
}

// ExtractZipcode 주소에서 우편번호 추출
func ExtractZipcode(address string) string {
	// 5자리 우편번호 패턴
	re := regexp.MustCompile(`\b\d{5}\b`)
	matches := re.FindString(address)
	return matches
}

// SplitAddress 주소를 구성 요소로 분리
func SplitAddress(address string) []string {
	// 공백으로 분리
	parts := strings.Fields(address)
	
	// 빈 요소 제거
	var result []string
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	
	return result
}