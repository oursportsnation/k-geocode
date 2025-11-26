package middleware

import (
	"fmt"
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// CORS Cross-Origin Resource Sharing 미들웨어
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 허용할 origin
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			// 기본값 (프로덕션에서는 특정 도메인만 허용하도록 수정 필요)
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		// 허용할 메서드
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		
		// 허용할 헤더
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Request-ID")
		
		// 자격 증명 허용
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		
		// 캐시 시간 (24시간)
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		
		// OPTIONS 요청 처리
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// CORSConfig CORS 설정
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig 기본 CORS 설정
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge: 12 * 60 * 60, // 12시간
	}
}

// CORSWithConfig 설정 기반 CORS 미들웨어
func CORSWithConfig(config CORSConfig) gin.HandlerFunc {
	// 기본값 설정
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = []string{"*"}
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{"GET", "POST"}
	}
	
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Origin 검증
		allowed := false
		for _, allowOrigin := range config.AllowOrigins {
			if allowOrigin == "*" || allowOrigin == origin {
				allowed = true
				break
			}
		}
		
		if allowed && origin != "" {
			// Origin 설정
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			
			// 자격 증명
			if config.AllowCredentials {
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		} else if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
			// 와일드카드
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		
		// Preflight 요청 처리
		if c.Request.Method == "OPTIONS" {
			// 메서드
			if len(config.AllowMethods) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(config.AllowMethods))
			}
			
			// 헤더
			if len(config.AllowHeaders) > 0 {
				c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders))
			}
			
			// 노출 헤더
			if len(config.ExposeHeaders) > 0 {
				c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders))
			}
			
			// 캐시
			if config.MaxAge > 0 {
				c.Writer.Header().Set("Access-Control-Max-Age", intToString(config.MaxAge))
			}
			
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		// 노출 헤더 (실제 요청)
		if len(config.ExposeHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders))
		}
		
		c.Next()
	}
}

// 헬퍼 함수
func joinStrings(strs []string) string {
	result := ""
	for i, str := range strs {
		if i > 0 {
			result += ", "
		}
		result += str
	}
	return result
}

func intToString(n int) string {
	return fmt.Sprintf("%d", n)
}