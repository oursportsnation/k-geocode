package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID Request ID 생성 및 추적 미들웨어
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 클라이언트가 보낸 Request ID 확인
		requestID := c.Request.Header.Get("X-Request-ID")
		
		// 없으면 새로 생성
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Context에 저장
		c.Set("requestID", requestID)
		
		// Response 헤더에 추가
		c.Writer.Header().Set("X-Request-ID", requestID)
		
		c.Next()
	}
}

// RequestIDWithConfig 설정 가능한 Request ID 미들웨어
type RequestIDConfig struct {
	// 헤더 이름
	HeaderName string
	// Context 키 이름
	ContextKey string
	// ID 생성 함수
	Generator func() string
}

// DefaultRequestIDConfig 기본 Request ID 설정
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		HeaderName: "X-Request-ID",
		ContextKey: "requestID",
		Generator: func() string {
			return uuid.New().String()
		},
	}
}

// RequestIDWithConfig 설정 기반 Request ID 미들웨어
func RequestIDWithConfig(config RequestIDConfig) gin.HandlerFunc {
	// 기본값 설정
	if config.HeaderName == "" {
		config.HeaderName = "X-Request-ID"
	}
	if config.ContextKey == "" {
		config.ContextKey = "requestID"
	}
	if config.Generator == nil {
		config.Generator = func() string {
			return uuid.New().String()
		}
	}
	
	return func(c *gin.Context) {
		// 클라이언트가 보낸 Request ID 확인
		requestID := c.Request.Header.Get(config.HeaderName)
		
		// 없으면 새로 생성
		if requestID == "" {
			requestID = config.Generator()
		}
		
		// Context에 저장
		c.Set(config.ContextKey, requestID)
		
		// Response 헤더에 추가
		c.Writer.Header().Set(config.HeaderName, requestID)
		
		c.Next()
	}
}

// GetRequestID Context에서 Request ID 가져오기
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("requestID"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// GetRequestIDWithKey 특정 키로 Request ID 가져오기
func GetRequestIDWithKey(c *gin.Context, key string) string {
	if requestID, exists := c.Get(key); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}