package provider

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeNotFound, "NOT_FOUND"},
		{ErrorTypeInvalid, "INVALID_INPUT"},
		{ErrorTypeSystemFailure, "SYSTEM_FAILURE"},
		{ErrorTypeTimeout, "TIMEOUT"},
		{ErrorTypeRateLimitExceeded, "RATE_LIMIT_EXCEEDED"},
		{ErrorTypeUnauthorized, "UNAUTHORIZED"},
		{ErrorType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.errorType.String())
		})
	}
}

func TestNewClassifiedError(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name        string
		errorType   ErrorType
		message     string
		wantRetry   bool
		wantFallback bool
	}{
		{
			name:        "NotFound allows fallback",
			errorType:   ErrorTypeNotFound,
			message:     "address not found",
			wantRetry:   true,
			wantFallback: true,
		},
		{
			name:        "Invalid prevents fallback",
			errorType:   ErrorTypeInvalid,
			message:     "invalid input",
			wantRetry:   false,
			wantFallback: false,
		},
		{
			name:        "SystemFailure allows fallback",
			errorType:   ErrorTypeSystemFailure,
			message:     "system error",
			wantRetry:   true,
			wantFallback: true,
		},
		{
			name:        "Timeout allows fallback",
			errorType:   ErrorTypeTimeout,
			message:     "request timeout",
			wantRetry:   true,
			wantFallback: true,
		},
		{
			name:        "RateLimitExceeded allows fallback",
			errorType:   ErrorTypeRateLimitExceeded,
			message:     "quota exceeded",
			wantRetry:   true,
			wantFallback: true,
		},
		{
			name:        "Unauthorized prevents fallback",
			errorType:   ErrorTypeUnauthorized,
			message:     "auth failed",
			wantRetry:   false,
			wantFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := NewClassifiedError(tt.errorType, tt.message, originalErr)

			require.NotNil(t, ce)
			assert.Equal(t, tt.errorType, ce.Type)
			assert.Equal(t, tt.message, ce.Message)
			assert.Equal(t, originalErr, ce.Original)
			assert.Equal(t, tt.wantRetry, ce.Retriable)
			assert.Equal(t, tt.wantFallback, ce.Fallback)
		})
	}
}

func TestClassifiedError_Error(t *testing.T) {
	originalErr := errors.New("connection refused")
	ce := NewClassifiedError(ErrorTypeSystemFailure, "failed to connect", originalErr)

	errStr := ce.Error()

	assert.Contains(t, errStr, "SYSTEM_FAILURE")
	assert.Contains(t, errStr, "failed to connect")
	assert.Contains(t, errStr, "connection refused")
}

func TestIsClassifiedError(t *testing.T) {
	t.Run("classified error", func(t *testing.T) {
		ce := NewClassifiedError(ErrorTypeNotFound, "not found", nil)

		result, ok := IsClassifiedError(ce)

		assert.True(t, ok)
		assert.Equal(t, ce, result)
	})

	t.Run("regular error", func(t *testing.T) {
		regularErr := errors.New("regular error")

		result, ok := IsClassifiedError(regularErr)

		assert.False(t, ok)
		assert.Nil(t, result)
	})

	t.Run("nil error", func(t *testing.T) {
		result, ok := IsClassifiedError(nil)

		assert.False(t, ok)
		assert.Nil(t, result)
	})
}

func TestPredefinedErrors(t *testing.T) {
	assert.NotNil(t, ErrAddressNotFound)
	assert.NotNil(t, ErrInvalidAddress)
	assert.NotNil(t, ErrAPIKeyInvalid)
	assert.NotNil(t, ErrQuotaExceeded)

	assert.Equal(t, "address not found", ErrAddressNotFound.Error())
	assert.Equal(t, "invalid address format", ErrInvalidAddress.Error())
	assert.Equal(t, "API key is invalid or expired", ErrAPIKeyInvalid.Error())
	assert.Equal(t, "daily quota exceeded", ErrQuotaExceeded.Error())
}

func TestDailyLimits(t *testing.T) {
	assert.Equal(t, 40000, DailyLimits["vWorld"])
	assert.Equal(t, 100000, DailyLimits["Kakao"])
}
