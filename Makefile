.PHONY: help setup build run test clean lint fmt dev test-unit test-integration test-coverage

# 기본 명령어
help:
	@echo "사용 가능한 명령어:"
	@echo "  make setup        - 개발 환경 설정"
	@echo "  make build        - 프로젝트 빌드"
	@echo "  make run          - 서버 실행"
	@echo "  make dev          - 개발 모드 실행 (hot reload)"
	@echo "  make test         - 테스트 실행"
	@echo "  make test-unit    - 단위 테스트만 실행"
	@echo "  make test-integration - 통합 테스트만 실행"
	@echo "  make test-coverage - 테스트 커버리지 확인"
	@echo "  make lint         - 코드 린트 검사"
	@echo "  make fmt          - 코드 포맷팅"
	@echo "  make clean        - 빌드 파일 삭제"

# 개발 환경 설정
setup:
	@echo "🔧 개발 환경 설정 중..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "✅ .env 파일 생성 완료 (API 키를 입력해주세요)"; fi
	@go mod download
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ 개발 환경 설정 완료!"

# 빌드
build:
	@echo "🔨 빌드 중..."
	@mkdir -p bin
	@go build -ldflags="-w -s" -o bin/geocoding-server cmd/server/main.go
	@echo "✅ 빌드 완료: bin/geocoding-server"

# 서버 실행
run:
	@echo "🚀 서버 시작 중..."
	@go run cmd/server/main.go

# 개발 모드 (hot reload)
dev:
	@echo "🔥 개발 모드 시작 중..."
	@which air > /dev/null || go install github.com/air-verse/air@latest
	@air

# 전체 테스트
test:
	@echo "🧪 테스트 실행 중..."
	@go test -v ./...

# 단위 테스트
test-unit:
	@echo "🧪 단위 테스트 실행 중..."
	@go test -v ./tests/unit/...

# 통합 테스트
test-integration:
	@echo "🧪 통합 테스트 실행 중..."
	@go test -v ./tests/integration/...

# 테스트 커버리지
test-coverage:
	@echo "📊 테스트 커버리지 확인 중..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 커버리지 리포트 생성: coverage.html"

# 린트
lint:
	@echo "🔍 코드 검사 중..."
	@golangci-lint run ./...

# 포맷팅
fmt:
	@echo "📝 코드 포맷팅 중..."
	@go fmt ./...
	@echo "✅ 포맷팅 완료"

# 정리
clean:
	@echo "🧹 정리 중..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "✅ 정리 완료"

# Swagger 문서 생성
swagger:
	@echo "📚 Swagger 문서 생성 중..."
	@swag init -g cmd/server/main.go -o docs
	@echo "✅ Swagger 문서 생성 완료"

# Docker 빌드 (추후 사용)
docker-build:
	@echo "🐳 Docker 이미지 빌드 중..."
	@docker build -t geocoding-service:latest .
	@echo "✅ Docker 이미지 빌드 완료"

# 환경 변수 체크
check-env:
	@echo "🔍 환경 변수 확인 중..."
	@if [ -z "${VWORLD_API_KEY}" ]; then echo "❌ VWORLD_API_KEY가 설정되지 않았습니다"; else echo "✅ VWORLD_API_KEY 설정됨"; fi
	@if [ -z "${KAKAO_API_KEY}" ]; then echo "❌ KAKAO_API_KEY가 설정되지 않았습니다"; else echo "✅ KAKAO_API_KEY 설정됨"; fi