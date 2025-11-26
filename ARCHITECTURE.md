# 시스템 아키텍처

## 개요

하이브리드 지오코딩 API 서버는 여러 무료 지오코딩 서비스를 활용하여 높은 가용성과 신뢰성을 제공하는 Go 기반 REST API 서버입니다.

## 아키텍처 다이어그램

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Client                         │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                    Gin Router                            │
│  • RequestID Middleware                                  │
│  • Logger Middleware                                     │
│  • Recovery Middleware                                   │
│  • CORS Middleware                                       │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                   Handlers                               │
│  • GeocodingHandler (단건/대량 지오코딩)                │
│  • HealthHandler (헬스체크)                              │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│               GeocodingService                           │
│  • 입력 검증 (주소 정규화, 유효성 검사)                  │
│  • Provider 폴백 로직                                    │
│  • 배치 처리 (동시성 제어)                               │
│  • 좌표 정규화 (소수점 6자리)                            │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
┌──────────────┐          ┌──────────────┐
│   vWorld     │          │    Kakao     │
│   Provider   │          │   Provider   │
│              │          │              │
│ • 도로명주소 │          │ • 유사검색   │
│ • 지번주소   │          │ • REST API   │
└──────────────┘          └──────────────┘
```

## 레이어 구조

### 1. HTTP Layer (`internal/handler`)
- **역할**: HTTP 요청/응답 처리
- **책임**:
  - 요청 파싱 및 검증
  - 응답 직렬화
  - HTTP 상태 코드 설정
  - 에러 응답 포맷팅

**주요 컴포넌트**:
- `GeocodingHandler`: 지오코딩 API 처리
- `HealthHandler`: 헬스체크 엔드포인트

### 2. Middleware Layer (`internal/middleware`)
- **역할**: 횡단 관심사(Cross-cutting concerns) 처리
- **구현된 미들웨어**:
  1. **RequestID**: 요청 추적을 위한 고유 ID 생성
  2. **Logger**: 구조화된 요청/응답 로깅
  3. **Recovery**: Panic 복구 및 500 에러 반환
  4. **CORS**: Cross-Origin 요청 처리

### 3. Service Layer (`internal/service`)
- **역할**: 비즈니스 로직 처리
- **책임**:
  - Provider 선택 및 폴백
  - 배치 처리 조율
  - 데이터 정규화
  - 에러 처리 및 분류

**주요 컴포넌트**:
- `GeocodingService`: 지오코딩 비즈니스 로직
- `Coordinator`: 서비스 초기화 및 관리

### 4. Provider Layer (`internal/provider`)
- **역할**: 외부 API 연동
- **책임**:
  - API 호출
  - 응답 파싱
  - 에러 처리 및 분류
  - 재시도 로직

**구현된 Provider**:
- `vWorld`: 국토지리정보원 API
- `Kakao`: 카카오 로컬 API

### 5. Utility Layer (`internal/utils`, `pkg`)
- **역할**: 공통 유틸리티 기능
- **모듈**:
  - `utils/math.go`: 좌표 계산 및 정규화
  - `utils/string.go`: 주소 정규화 및 검증
  - `pkg/logger`: 로깅 추상화
  - `pkg/httpclient`: HTTP 클라이언트 최적화

## 핵심 메커니즘

### 1. 폴백 메커니즘

```go
// 의사 코드
for each provider in [vWorld, Kakao] {
    if !provider.IsAvailable() {
        continue  // 사용 불가능한 Provider 건너뛰기
    }

    result, err := provider.Geocode(address)

    if err != nil {
        if err.Type == UNAUTHORIZED || err.Type == INVALID {
            break  // 재시도 불가능한 에러
        }
        continue  // 다음 Provider 시도
    }

    if result.Success {
        return result  // 성공
    }
}

return "모든 Provider 실패"
```

**폴백 순서**:
1. vWorld (1순위)
   - 도로명 주소 조회
   - 실패 시 지번 주소로 재시도
2. Kakao (2순위)
   - 유사 검색으로 조회

### 2. 배치 처리

```go
// 동시성 제어
maxConcurrent := 10
semaphore := make(chan struct{}, maxConcurrent)

var wg sync.WaitGroup
results := make([]*GeocodingResponse, len(addresses))

for i, address := range addresses {
    wg.Add(1)
    semaphore <- struct{}{}  // 동시 실행 제한

    go func(index int, addr string) {
        defer wg.Done()
        defer func() { <-semaphore }()

        results[index] = geocode(addr)
    }(i, address)
}

wg.Wait()
return results
```

**특징**:
- 최대 10개 동시 처리
- 세마포어 패턴으로 동시성 제어
- 개별 실패가 전체에 영향 없음

### 3. 에러 분류 시스템

```go
type ErrorType int

const (
    NotFound        ErrorType = iota  // 폴백 허용
    SystemFailure                     // 폴백 허용
    Timeout                           // 폴백 허용
    RateLimitExceeded                 // 폴백 허용
    Invalid                           // 폴백 불가
    Unauthorized                      // 폴백 불가
)
```

**폴백 결정**:
- **허용**: NotFound, SystemFailure, Timeout, RateLimitExceeded
- **불가**: Invalid (잘못된 입력), Unauthorized (인증 실패)

### 4. 좌표 정규화

```go
// 소수점 6자리로 정규화 (약 0.1m 정밀도)
func RoundToSixDecimal(value float64) float64 {
    return math.Round(value * 1000000) / 1000000
}
```

**정밀도**:
- 소수점 6자리 = 약 0.1m 오차
- WGS84 좌표계 (EPSG:4326)

## 디렉토리 구조

```
geocoding-service/
├── cmd/
│   └── server/
│       └── main.go              # 서버 엔트리포인트
├── internal/
│   ├── config/
│   │   └── config.go            # 설정 로더
│   ├── handler/
│   │   ├── geocoding.go         # 지오코딩 핸들러
│   │   └── health.go            # 헬스체크 핸들러
│   ├── middleware/
│   │   ├── logger.go            # 로깅 미들웨어
│   │   ├── recovery.go          # Panic 복구
│   │   ├── requestid.go         # Request ID
│   │   └── cors.go              # CORS
│   ├── model/
│   │   └── geocoding.go         # 데이터 모델
│   ├── provider/
│   │   ├── provider.go          # Provider 인터페이스
│   │   ├── errors.go            # 에러 분류
│   │   ├── vworld.go            # vWorld Provider
│   │   └── kakao.go             # Kakao Provider
│   ├── service/
│   │   ├── geocoding.go         # 지오코딩 서비스
│   │   └── coordinator.go       # 서비스 조율자
│   └── utils/
│       ├── math.go              # 수학 유틸리티
│       └── string.go            # 문자열 유틸리티
├── pkg/
│   ├── logger/
│   │   └── zap.go               # Logger 래퍼
│   └── httpclient/
│       └── client.go            # HTTP 클라이언트
├── configs/
│   └── config.yaml              # 설정 파일
├── tests/
│   ├── unit/                    # 단위 테스트
│   └── integration/             # 통합 테스트
└── docs/                        # 문서
```

## 의존성 주입

인터페이스 기반 설계로 테스트 가능성과 유연성을 확보:

```go
// Service는 인터페이스에 의존
type GeocodingServiceInterface interface {
    Geocode(ctx context.Context, address string) (*GeocodingResponse, error)
    GeocodeBatch(ctx context.Context, addresses []string) (*BulkResponse, error)
}

// Handler는 인터페이스를 주입받음
type GeocodingHandler struct {
    service GeocodingServiceInterface
    logger  *zap.Logger
}
```

**장점**:
- 단위 테스트에서 Mock 사용 가능
- 구현체 교체 용이
- 순환 참조 방지

## 설정 관리

### 다층 설정 시스템

```yaml
# configs/config.yaml
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 30s

providers:
  vworld:
    api_url: "https://api.vworld.kr/req/address"
    timeout: 5s
  kakao:
    api_url: "https://dapi.kakao.com/v2/local/search/address.json"
    timeout: 5s
```

**환경변수 오버라이드**:
```bash
VWORLD_API_KEY=your_key
KAKAO_API_KEY=your_key
SERVER_PORT=8080
```

## 로깅 전략

### 구조화된 로깅 (Uber Zap)

```go
logger.Info("Geocoding request received",
    zap.String("request_id", requestID),
    zap.String("address", address),
)
```

**로그 레벨**:
- **DEBUG**: Provider 시도, 재시도 정보
- **INFO**: 요청/응답, 성공 케이스
- **WARN**: 폴백, 재시도, 일시적 실패
- **ERROR**: 심각한 에러, 전체 실패

**로그 필드**:
- `request_id`: 요청 추적
- `provider`: Provider 이름
- `error_type`: 에러 분류
- `duration`: 처리 시간

## 성능 최적화

### 1. HTTP 클라이언트 풀링

```go
Transport: &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
```

### 2. 동시성 제어
- 세마포어 패턴으로 최대 10개 동시 요청
- 고루틴 재사용으로 메모리 효율화

### 3. 좌표 정규화
- 소수점 6자리로 통일하여 응답 크기 최소화

## 확장성

### Phase 2 계획
1. **Circuit Breaker**: Provider 장애 격리
2. **Rate Limiting**: 사용량 제한
3. **Caching**: Redis 기반 캐싱
4. **Monitoring**: Prometheus 메트릭

### 새로운 Provider 추가

```go
// 1. Provider 인터페이스 구현
type NewProvider struct {
    apiKey string
    client *http.Client
}

func (p *NewProvider) Geocode(ctx context.Context, address string) (*ProviderResult, error) {
    // 구현
}

// 2. Coordinator에 등록
coordinator.addProvider(newProvider)
```

## 테스트 전략

### 단위 테스트
- Provider: API 응답 파싱
- Service: 폴백 로직
- Utils: 좌표 계산, 주소 검증
- Handler: HTTP 요청/응답

### 통합 테스트
- End-to-end API 테스트
- 폴백 시나리오 검증
- 에러 처리 검증

**테스트 커버리지**: 핵심 비즈니스 로직 100%

## 보안 고려사항

1. **API 키 관리**: 환경변수로 관리, 코드에 하드코딩 금지
2. **입력 검증**: 모든 사용자 입력 검증
3. **에러 메시지**: 민감한 정보 노출 방지
4. **Rate Limiting**: (Phase 2) DoS 공격 방어
5. **CORS**: 허용된 Origin만 접근 가능

## 운영 고려사항

### Health Check
- `/ping`: 서버 생존 확인
- `/health`: Provider 상태 확인
- `/ready`: Kubernetes Readiness Probe

### Graceful Shutdown
```go
// SIGTERM/SIGINT 처리
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
server.Shutdown(ctx)
```

### 모니터링 지표 (Phase 2)
- 요청 수 (성공/실패)
- 응답 시간
- Provider별 성공률
- 에러율

---

**작성일**: 2025-11-24
**버전**: Phase 1 완료
