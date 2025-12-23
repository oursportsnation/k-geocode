package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// CORS Tests
func TestCORS(t *testing.T) {
	router := setupTestRouter()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	t.Run("GET request with Origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("GET request without Origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	assert.Equal(t, []string{"*"}, config.AllowOrigins)
	assert.Contains(t, config.AllowMethods, "GET")
	assert.Contains(t, config.AllowMethods, "POST")
	assert.Contains(t, config.AllowHeaders, "Content-Type")
	assert.Equal(t, 12*60*60, config.MaxAge)
}

func TestCORSWithConfig(t *testing.T) {
	t.Run("with custom config", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:     []string{"http://allowed.com"},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"X-Custom-Header"},
			AllowCredentials: true,
			MaxAge:           3600,
		}

		router := setupTestRouter()
		router.Use(CORSWithConfig(config))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://allowed.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://allowed.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("with empty config uses defaults", func(t *testing.T) {
		config := CORSConfig{}

		router := setupTestRouter()
		router.Use(CORSWithConfig(config))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("OPTIONS with expose headers", func(t *testing.T) {
		config := CORSConfig{
			AllowOrigins:  []string{"*"},
			AllowMethods:  []string{"GET"},
			ExposeHeaders: []string{"X-Total-Count"},
			MaxAge:        3600,
		}

		router := setupTestRouter()
		router.Use(CORSWithConfig(config))
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "X-Total-Count", w.Header().Get("Access-Control-Expose-Headers"))
		assert.Equal(t, "3600", w.Header().Get("Access-Control-Max-Age"))
	})
}

// RequestID Tests
func TestRequestID(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())

	var capturedID string
	router.GET("/test", func(c *gin.Context) {
		capturedID = c.GetString("requestID")
		c.String(http.StatusOK, "OK")
	})

	t.Run("generates new ID when not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, capturedID)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	})

	t.Run("uses provided ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "custom-id-123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "custom-id-123", capturedID)
		assert.Equal(t, "custom-id-123", w.Header().Get("X-Request-ID"))
	})
}

func TestDefaultRequestIDConfig(t *testing.T) {
	config := DefaultRequestIDConfig()

	assert.Equal(t, "X-Request-ID", config.HeaderName)
	assert.Equal(t, "requestID", config.ContextKey)
	assert.NotNil(t, config.Generator)

	// Generator should produce valid UUIDs
	id := config.Generator()
	assert.NotEmpty(t, id)
	assert.Len(t, id, 36) // UUID format
}

func TestRequestIDWithConfig(t *testing.T) {
	t.Run("with custom config", func(t *testing.T) {
		config := RequestIDConfig{
			HeaderName: "X-Custom-ID",
			ContextKey: "customID",
			Generator: func() string {
				return "generated-id"
			},
		}

		router := setupTestRouter()
		router.Use(RequestIDWithConfig(config))

		var capturedID string
		router.GET("/test", func(c *gin.Context) {
			capturedID = c.GetString("customID")
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, "generated-id", capturedID)
		assert.Equal(t, "generated-id", w.Header().Get("X-Custom-ID"))
	})

	t.Run("with empty config uses defaults", func(t *testing.T) {
		config := RequestIDConfig{}

		router := setupTestRouter()
		router.Use(RequestIDWithConfig(config))

		var capturedID string
		router.GET("/test", func(c *gin.Context) {
			capturedID = c.GetString("requestID")
			c.String(http.StatusOK, "OK")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.NotEmpty(t, capturedID)
	})
}

func TestGetRequestID(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())

	var result string
	router.GET("/test", func(c *gin.Context) {
		result = GetRequestID(c)
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.NotEmpty(t, result)
}

func TestGetRequestIDWithKey(t *testing.T) {
	router := setupTestRouter()

	router.GET("/test", func(c *gin.Context) {
		c.Set("myKey", "myValue")
		result := GetRequestIDWithKey(c, "myKey")
		assert.Equal(t, "myValue", result)

		// Non-existent key
		result = GetRequestIDWithKey(c, "nonExistent")
		assert.Empty(t, result)

		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Logger Tests
func TestLogger(t *testing.T) {
	logger := zap.NewNop()
	router := setupTestRouter()
	router.Use(Logger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogger_StatusCodes(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		statusCode int
	}{
		{"success", http.StatusOK},
		{"client error", http.StatusBadRequest},
		{"server error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()
			router.Use(Logger(logger))
			router.GET("/test", func(c *gin.Context) {
				c.String(tt.statusCode, "response")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestLogger_WithErrors(t *testing.T) {
	logger := zap.NewNop()
	router := setupTestRouter()
	router.Use(Logger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.Error(gin.Error{Err: assert.AnError, Type: gin.ErrorTypePublic})
		c.String(http.StatusBadRequest, "error")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGinLogger(t *testing.T) {
	logger := zap.NewNop()
	router := setupTestRouter()
	router.Use(GinLogger(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Recovery Tests
func TestRecovery(t *testing.T) {
	logger := zap.NewNop()
	router := setupTestRouter()
	router.Use(Recovery(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	// Should not panic
	require.NotPanics(t, func() {
		router.ServeHTTP(w, req)
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecovery_NoPanic(t *testing.T) {
	logger := zap.NewNop()
	router := setupTestRouter()
	router.Use(Recovery(logger))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomRecovery(t *testing.T) {
	logger := zap.NewNop()
	customHandlerCalled := false

	router := setupTestRouter()
	router.Use(CustomRecovery(logger, func(c *gin.Context, recovered interface{}) {
		customHandlerCalled = true
		c.JSON(http.StatusInternalServerError, gin.H{"custom": "error"})
		c.Abort()
	}))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	require.NotPanics(t, func() {
		router.ServeHTTP(w, req)
	})

	assert.True(t, customHandlerCalled)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCustomRecovery_NilHandler(t *testing.T) {
	logger := zap.NewNop()

	router := setupTestRouter()
	router.Use(CustomRecovery(logger, nil))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	require.NotPanics(t, func() {
		router.ServeHTTP(w, req)
	})

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Helper function tests
func TestJoinStrings(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"a", "b", "c"}, "a, b, c"},
		{[]string{"single"}, "single"},
		{[]string{}, ""},
	}

	for _, tt := range tests {
		result := joinStrings(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{123, "123"},
		{-456, "-456"},
	}

	for _, tt := range tests {
		result := intToString(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
