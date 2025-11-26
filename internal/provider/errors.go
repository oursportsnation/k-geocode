package provider

import (
	"errors"
	"fmt"
)

// ErrorType 에러 분류
type ErrorType int

const (
	ErrorTypeNotFound          ErrorType = iota // 주소를 찾을 수 없음 (폴백 가능)
	ErrorTypeInvalid                            // 잘못된 입력 형식 (폴백 불가)
	ErrorTypeSystemFailure                      // 시스템 오류 (폴백 가능)
	ErrorTypeTimeout                            // 요청 타임아웃 (폴백 가능)
	ErrorTypeRateLimitExceeded                  // 일일 할당량 초과 (폴백 가능)
	ErrorTypeUnauthorized                       // 인증 실패 (폴백 불가)
)

// ClassifiedError 분류된 에러
type ClassifiedError struct {
	Type      ErrorType
	Message   string
	Original  error
	Retriable bool // 재시도 가능 여부
	Fallback  bool // 다음 Provider로 폴백 가능 여부
}

func (ce *ClassifiedError) Error() string {
	return fmt.Sprintf("[%s] %s: %v", ce.Type.String(), ce.Message, ce.Original)
}

func (t ErrorType) String() string {
	switch t {
	case ErrorTypeNotFound:
		return "NOT_FOUND"
	case ErrorTypeInvalid:
		return "INVALID_INPUT"
	case ErrorTypeSystemFailure:
		return "SYSTEM_FAILURE"
	case ErrorTypeTimeout:
		return "TIMEOUT"
	case ErrorTypeRateLimitExceeded:
		return "RATE_LIMIT_EXCEEDED"
	case ErrorTypeUnauthorized:
		return "UNAUTHORIZED"
	default:
		return "UNKNOWN"
	}
}

// NewClassifiedError 에러 분류 생성자
func NewClassifiedError(errorType ErrorType, message string, original error) *ClassifiedError {
	ce := &ClassifiedError{
		Type:     errorType,
		Message:  message,
		Original: original,
	}
	
	// 에러 타입별 재시도/폴백 여부 결정
	switch errorType {
	case ErrorTypeNotFound, ErrorTypeSystemFailure, ErrorTypeTimeout, ErrorTypeRateLimitExceeded:
		ce.Retriable = true
		ce.Fallback = true
	case ErrorTypeInvalid, ErrorTypeUnauthorized:
		ce.Retriable = false
		ce.Fallback = false
	}
	
	return ce
}

// IsClassifiedError 분류된 에러인지 확인
func IsClassifiedError(err error) (*ClassifiedError, bool) {
	ce, ok := err.(*ClassifiedError)
	return ce, ok
}

// 일반적인 에러들
var (
	ErrAddressNotFound = errors.New("address not found")
	ErrInvalidAddress  = errors.New("invalid address format")
	ErrAPIKeyInvalid   = errors.New("API key is invalid or expired")
	ErrQuotaExceeded   = errors.New("daily quota exceeded")
)