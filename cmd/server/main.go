// Copyright 2025 Our Sports Nation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	geocoding "github.com/oursportsnation/k-geocode"
	"github.com/oursportsnation/k-geocode/internal/config"
	"github.com/oursportsnation/k-geocode/internal/handler"
	"github.com/oursportsnation/k-geocode/internal/middleware"
	"github.com/oursportsnation/k-geocode/internal/service"
	"github.com/oursportsnation/k-geocode/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "github.com/oursportsnation/k-geocode/docs" // Swagger docs
)

// @title           하이브리드 지오코딩 API
// @version         1.0
// @description     한글 주소를 WGS84 좌표로 변환하는 하이브리드 지오코딩 API 서버입니다.
// @description     vWorld와 Kakao API를 활용한 자동 폴백 시스템을 제공합니다.

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /

// @schemes   http https

// @tag.name geocoding
// @tag.description 지오코딩 API
// @tag.name health
// @tag.description 헬스체크 API

func main() {
	// .env 파일 로드 (있으면)
	if err := godotenv.Load(); err != nil {
		// .env 파일이 없어도 계속 진행
		log.Println("No .env file found")
	}

	// 환경 설정
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	// 설정 파일 로드
	configPath := "configs/config.yaml"
	cfg, err := config.LoadWithEnv(configPath, env)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Logger 초기화
	appLogger, err := logger.New(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// 시작 로그
	appLogger.Info("Starting Geocoding Service",
		zap.String("port", cfg.Server.Port),
		zap.String("environment", env),
		zap.String("log_level", cfg.Logging.Level),
	)

	// Gin 모드 설정
	if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Coordinator 설정 (Provider 초기화 포함)
	coordinator, err := service.NewCoordinator(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to create coordinator", zap.Error(err))
	}

	// Service 설정
	geocodingService := coordinator.GetGeocodingService()

	// Router 설정
	router := setupRouter(cfg, geocodingService, coordinator, appLogger)

	// 서버 설정
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown 설정
	go func() {
		// 서비스 시작
		appLogger.Info("Server started", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 서버 시작 정보 출력 (클릭 가능한 링크)
	printStartupBanner(cfg.Server.Port)

	// 시그널 대기
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// 5초 타임아웃으로 graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Server exiting")
}


// setupRouter Router 설정
func setupRouter(cfg *config.Config, geocodingService *service.GeocodingService, coordinator *service.Coordinator, logger *zap.Logger) *gin.Engine {
	router := gin.New()

	// 미들웨어 설정
	router.Use(middleware.RequestID())                    // Request ID (먼저 설정)
	router.Use(middleware.Logger(logger))                 // 로깅
	router.Use(middleware.Recovery(logger))               // 패닉 리커버리
	router.Use(middleware.CORS())                         // CORS

	// 핸들러 생성
	geocodingHandler := handler.NewGeocodingHandler(geocodingService, logger)
	healthHandler := handler.NewHealthHandler(coordinator, logger)

	// Swagger 문서
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 헬스체크 라우트
	router.GET("/ping", healthHandler.Ping)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API v1 라우트 그룹
	v1 := router.Group("/api/v1")
	{
		// 지오코딩 API
		v1.POST("/geocode", geocodingHandler.Geocode)
		v1.POST("/geocode/bulk", geocodingHandler.GeocodeBulk)
	}

	// 404 핸들러
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "not found",
			"request_id": middleware.GetRequestID(c),
		})
	})

	return router
}

// printStartupBanner 서버 시작 배너 출력
func printStartupBanner(port string) {
	fmt.Println()
	fmt.Println("=============================================================")
	fmt.Printf("              k-geocode Server v%s\n", geocoding.Version)
	fmt.Println("=============================================================")
	fmt.Println()
	fmt.Printf("  Server:      http://localhost:%s\n", port)
	fmt.Printf("  Swagger UI:  http://localhost:%s/swagger/index.html\n", port)
	fmt.Printf("  Health:      http://localhost:%s/health\n", port)
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println("=============================================================")
	fmt.Println()
}
