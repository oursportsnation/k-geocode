# Basic Usage Example

이 예제는 k-geocode를 Go 프로젝트에서 라이브러리로 사용하는 기본적인 방법을 보여줍니다.

## 실행 방법

### 1. 환경 변수 설정

```bash
export VWORLD_API_KEY="your-vworld-api-key"
export KAKAO_API_KEY="your-kakao-api-key"
```

### 2. 예제 실행

```bash
cd examples/basic
go run main.go
```

## 예제 내용

이 예제는 다음 기능들을 시연합니다:

1. **클라이언트 생성** - Config를 사용한 클라이언트 초기화
2. **Provider 확인** - 사용 가능한 Provider 목록 조회
3. **단일 주소 지오코딩** - 기본 지오코딩 요청
4. **타입 지정 지오코딩** - ROAD/PARCEL 타입을 지정한 지오코딩
5. **배치 지오코딩** - 여러 주소를 한 번에 처리
6. **서비스 가용성 확인** - 서비스 상태 체크

## 출력 예시

```
Available providers: [vWorld Kakao]

=== Single Address Geocoding ===
Address: 서울특별시 중구 세종대로 110
Coordinates: (37.566535, 126.977969)
Provider: vWorld
Road Address: 서울특별시 중구 세종대로 110
Parcel Address: 서울특별시 중구 태평로1가 31

Provider Attempts:
  ✓ vWorld

=== Geocoding with Address Type ===
Address: 서울특별시 강남구 테헤란로 152
Coordinates: (37.505228, 127.052487)
Provider: vWorld

=== Batch Geocoding ===
Processed 3 addresses:
1. 서울특별시 중구 세종대로 110 -> (37.566535, 126.977969) [vWorld]
2. 부산광역시 해운대구 해운대해변로 264 -> (35.158698, 129.160384) [vWorld]
3. 대구광역시 중구 공평로 88 -> (35.871435, 128.597344) [vWorld]

=== Service Availability ===
Service is available ✓
```

## 참고사항

- 최소 하나의 API 키(vWorld 또는 Kakao)가 필요합니다
- 기본 타임아웃은 10초로 설정되어 있습니다
- 배치 처리는 최대 100개까지 가능합니다
- 자동 폴백 기능으로 vWorld 실패 시 Kakao로 재시도합니다
