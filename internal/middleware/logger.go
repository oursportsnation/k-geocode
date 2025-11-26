package middleware

import (
	"time"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger Gin 로깅 미들웨어
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 시작 시간
		start := time.Now()
		
		// Request ID (다른 미들웨어에서 설정)
		requestID := c.GetString("requestID")
		
		// 요청 로깅
		logger.Info("incoming request",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
		
		// 다음 핸들러 실행
		c.Next()
		
		// 응답 시간 계산
		latency := time.Since(start)
		
		// 응답 로깅
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		}
		
		// 에러가 있으면 추가
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}
		
		// 상태 코드에 따라 로그 레벨 결정
		statusCode := c.Writer.Status()
		switch {
		case statusCode >= 500:
			logger.Error("request completed", fields...)
		case statusCode >= 400:
			logger.Warn("request completed", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	}
}

// GinLogger Gin의 기본 로거를 Zap으로 대체
func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(params gin.LogFormatterParams) string {
			// Zap 로거 사용
			logger.Info("gin request",
				zap.String("client_ip", params.ClientIP),
				zap.String("method", params.Method),
				zap.String("path", params.Path),
				zap.Int("status", params.StatusCode),
				zap.Duration("latency", params.Latency),
				zap.String("error", params.ErrorMessage),
			)
			return ""
		},
		Output:    nil, // Zap이 처리하므로 nil
		SkipPaths: []string{"/health", "/ping"}, // 헬스체크는 로깅 제외
	})
}