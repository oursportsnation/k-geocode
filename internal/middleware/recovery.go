package middleware

import (
	"net/http"
	"runtime/debug"
	
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery panic 리커버리 미들웨어
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 스택 트레이스 가져오기
				stack := debug.Stack()
				
				// Request ID
				requestID := c.GetString("requestID")
				
				// 로그 기록
				logger.Error("panic recovered",
					zap.String("request_id", requestID),
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("stack", string(stack)),
				)
				
				// Connection 리셋 확인
				if c.IsAborted() {
					return
				}
				
				// 에러 응답
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":      "internal server error",
					"request_id": requestID,
				})
				
				// 후속 처리 중단
				c.Abort()
			}
		}()
		
		c.Next()
	}
}

// CustomRecovery 커스텀 리커버리 핸들러를 사용하는 미들웨어
func CustomRecovery(logger *zap.Logger, handle gin.RecoveryFunc) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Request ID
		requestID := c.GetString("requestID")
		
		// 로그 기록
		logger.Error("panic recovered with custom handler",
			zap.String("request_id", requestID),
			zap.Any("error", recovered),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)
		
		// 커스텀 핸들러 호출
		if handle != nil {
			handle(c, recovered)
		} else {
			// 기본 응답
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":      "internal server error",
				"request_id": requestID,
			})
			c.Abort()
		}
	})
}