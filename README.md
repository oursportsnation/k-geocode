# k-geocode

[![Go Reference](https://pkg.go.dev/badge/github.com/oursportsnation/k-geocode.svg)](https://pkg.go.dev/github.com/oursportsnation/k-geocode)
[![Go Report Card](https://goreportcard.com/badge/github.com/oursportsnation/k-geocode)](https://goreportcard.com/report/github.com/oursportsnation/k-geocode)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

한국 주소를 WGS84 좌표로 변환하는 하이브리드 지오코딩 Go 패키지입니다.

## ✨ 주요 특징

- 🔄 **자동 폴백**: vWorld → Kakao 순차 시도로 높은 성공률
- 📍 **정밀한 좌표**: 소수점 6자리 (약 0.1m 정밀도)
- ⚡ **배치 처리**: 최대 100건 동시 처리 지원
- 🎯 **주소 타입 지정**: ROAD(도로명) 또는 PARCEL(지번) 선택 가능
- 🛡️ **안정성**: 에러 분류 및 재시도 로직
- 🔍 **투명한 디버깅**: 모든 Provider 시도 내역 추적
- 📊 **모니터링**: 구조화된 로깅 및 헬스체크
- 📚 **Swagger UI**: 대화형 API 문서 (`/swagger/index.html`)
- 🧪 **테스트**: 85개+ 단위 테스트 (46.8% 커버리지)

## 📦 설치

```bash
go get github.com/oursportsnation/k-geocode
```

## 🚀 빠른 시작

### 사전 요구사항

- Go 1.21 이상
- vWorld API 키 ([발급 링크](https://www.vworld.kr/dev/v4dv_apiDevice2_s001.do))
- Kakao REST API 키 ([발급 링크](https://developers.kakao.com/))

### Go 패키지로 사용 (권장)

```go
package main

import (
    "context"
    "log"

    geocoding "github.com/oursportsnation/k-geocode"
)

func main() {
    // 지오코딩 클라이언트 생성
    client, err := geocoding.New(geocoding.Config{
        VWorldAPIKey: "your-vworld-key",
        KakaoAPIKey:  "your-kakao-key",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 주소 → 좌표 변환
    result, err := client.Geocode(context.Background(), "서울시 강남구 테헤란로 152")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("위도: %f, 경도: %f", result.Latitude, result.Longitude)
}
```

더 많은 예제는 **[examples/basic](./examples/basic)**를 참고하세요.

### 독립 서버로 실행

REST API 서버로 실행할 수도 있습니다:

```bash
# 1. 저장소 클론
git clone <repository-url>
cd geocoding-service

# 2. 의존성 설치
go mod download

# 3. 환경 설정
cp .env.example .env
# .env 파일을 열어 API 키 입력

# 4. 서버 실행
make run
```

서버가 `http://localhost:8080` 에서 실행됩니다.

## 📖 API 사용법

### 단건 지오코딩

```bash
# 기본 요청 (자동 폴백: ROAD → PARCEL)
curl -X POST http://localhost:8080/api/v1/geocode \
  -H "Content-Type: application/json" \
  -d '{"address": "서울특별시 중구 세종대로 110"}'

# 주소 타입 지정 (ROAD 또는 PARCEL)
curl -X POST http://localhost:8080/api/v1/geocode \
  -H "Content-Type: application/json" \
  -d '{"address": "서울특별시 중구 세종대로 110", "address_type": "ROAD"}'
```

**응답 예시**:
```json
{
  "success": true,
  "coordinate": {
    "latitude": 37.566826,
    "longitude": 126.978652
  },
  "address_detail": {
    "road_address": "서울특별시 중구 세종대로 110",
    "parcel_address": "서울특별시 중구 태평로1가 31",
    "building_name": "서울시청"
  },
  "provider": "vWorld",
  "attempts": [
    {
      "provider": "vWorld",
      "success": true
    }
  ],
  "processed_at": "2025-11-25T10:00:00.000000+09:00",
  "processing_time_ms": 123
}
```

### 대량 지오코딩

```bash
curl -X POST http://localhost:8080/api/v1/geocode/bulk \
  -H "Content-Type: application/json" \
  -d '{
    "addresses": [
      "서울시 강남구",
      "서울시 서초구"
    ]
  }'
```

**응답 예시**:
```json
{
  "results": [
    {
      "success": true,
      "coordinate": {"latitude": 37.517331, "longitude": 127.047374},
      "address_detail": {
        "road_address": "서울특별시 강남구"
      },
      "provider": "vWorld",
      "attempts": [
        {
          "provider": "vWorld",
          "success": true
        }
      ],
      "processed_at": "2025-11-25T10:00:00Z",
      "processing_time_ms": 32
    },
    {
      "success": true,
      "coordinate": {"latitude": 37.483569, "longitude": 127.032598},
      "address_detail": {
        "road_address": "서울특별시 서초구"
      },
      "provider": "vWorld",
      "attempts": [
        {
          "provider": "vWorld",
          "success": true
        }
      ],
      "processed_at": "2025-11-25T10:00:01Z",
      "processing_time_ms": 28
    }
  ],
  "summary": {
    "total": 2,
    "success": 2,
    "failed": 0
  },
  "processing_time_ms": 95
}
```

### 헬스 체크

```bash
# 간단한 Ping
curl http://localhost:8080/ping

# 상세 헬스 체크
curl http://localhost:8080/health

# Readiness Probe
curl http://localhost:8080/ready
```

자세한 API 문서는 [API.md](./API.md)를 참고하세요.

## 🧪 테스트

```bash
# 단위 테스트 실행
make test

# 테스트 커버리지
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 통합 테스트 실행
./tests/integration/api_test.sh
```

### 커버리지 현황

| 패키지 | 커버리지 |
|--------|----------|
| internal/utils | 100.0% |
| pkg/httpclient | 100.0% |
| internal/middleware | 96.6% |
| pkg/logger | 95.5% |
| internal/handler | 94.9% |
| (root) geocoding | 70.9% |
| internal/service | 64.2% |
| **전체** | **46.8%** |

## 📚 문서

- **[API.md](./API.md)** - API 레퍼런스
- **[examples/basic](./examples/basic)** - 기본 사용 예제
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - 시스템 아키텍처 및 설계
- **[docs/development-history.md](./docs/development-history.md)** - 개발 히스토리
- **[docs/testing.md](./docs/testing.md)** - 테스트 가이드
- **[docs/implementation-plan.md](./docs/implementation-plan.md)** - 상세 구현 계획
- **[docs/original-spec.md](./docs/original-spec.md)** - 원본 기술 백서

## 🏗️ 프로젝트 구조

```
├── client.go           # 공개 API 클라이언트
├── config.go           # 공개 설정 구조체
├── types.go            # 공개 타입 정의
├── cmd/server/         # 서버 엔트리포인트
├── internal/
│   ├── handler/        # HTTP 핸들러
│   ├── middleware/     # 미들웨어
│   ├── service/        # 비즈니스 로직
│   ├── provider/       # 외부 API 연동
│   └── utils/          # 유틸리티
├── pkg/                # 공용 패키지
├── examples/           # 사용 예제
│   └── basic/          # 기본 사용 예제
├── configs/            # 설정 파일
├── tests/              # 테스트
└── docs/               # 문서
```

자세한 내용은 [ARCHITECTURE.md](./ARCHITECTURE.md)를 참고하세요.

## 🔧 개발 도구

```bash
# 코드 빌드
make build

# 서버 실행
make run

# 테스트 실행
make test

# 코드 정리
make clean
```

## 🛠️ 설정

### 환경변수

`.env` 파일에 다음 변수를 설정하세요:

```bash
# API 키
VWORLD_API_KEY=your_vworld_api_key
KAKAO_API_KEY=your_kakao_rest_api_key

# 서버 설정 (선택사항)
SERVER_PORT=8080
LOG_LEVEL=info
```

### 설정 파일

`configs/config.yaml`에서 상세 설정을 조정할 수 있습니다:

```yaml
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 30s

providers:
  vworld:
    timeout: 5s
  kakao:
    timeout: 5s
```

## 📊 현재 상태

**v0.1.0** (2025-12-23)

- ✅ 핵심 기능 구현
- ✅ 단위 테스트 85개+ (46.8% 커버리지)
- ✅ 통합 테스트 16개
- ✅ 폴백 메커니즘 검증
- ✅ 배치 처리 구현
- ✅ Swagger/OpenAPI 문서화
- ✅ Provider 시도 내역 추적
- ✅ 주소 타입 지정 기능 (ROAD/PARCEL)
- ✅ vWorld API 버그 수정
- ✅ **Go 패키지** - `go get`으로 설치 가능
- ✅ godoc 스타일 문서화

**계획 중**

- ⏳ Circuit Breaker 구현
- ⏳ Redis 캐싱
- ⏳ Rate Limiting
- ⏳ Prometheus 메트릭

## 🤝 기여

기여는 언제나 환영합니다! Pull Request를 보내주세요.

## 📄 라이선스

Apache License 2.0

이 프로젝트는 Apache License 2.0에 따라 라이선스가 부여됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 🙋 문의

이슈가 있으시면 GitHub Issues에 등록해주세요.

---

**최종 업데이트**: 2025-12-23
**버전**: v0.1.0
