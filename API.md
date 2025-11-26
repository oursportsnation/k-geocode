# Geocoding Service API Documentation

## Base URL
```
http://localhost:8080
```

## Endpoints

### 1. Health Check

#### GET /ping
Simple ping endpoint to check if the service is running.

**Response:**
```json
{
    "message": "pong",
    "time": "2025-11-20T17:44:12+09:00"
}
```

#### GET /health
Detailed health status including provider availability and system metrics.

**Response:**
```json
{
    "status": "healthy",
    "timestamp": "2025-11-20T17:44:30.132726+09:00",
    "providers": [
        {
            "name": "vWorld",
            "available": true
        },
        {
            "name": "Kakao",
            "available": true
        }
    ],
    "system": {
        "uptime": "33.848874125s",
        "goroutines": 5,
        "memory_mb": 0.840728759765625,
        "num_gc": 0
    }
}
```

#### GET /ready
Check if the service is ready to handle requests.

**Response:**
```json
{
    "ready": true
}
```

### 2. Geocoding

#### POST /api/v1/geocode
Convert a single Korean address to WGS84 coordinates.

**Request:**
```json
{
    "address": "서울특별시 중구 세종대로 110",
    "address_type": "ROAD"  // Optional: "ROAD" or "PARCEL" (자동 폴백 시 생략 가능)
}
```

**Parameters:**
- `address` (required): 한글 주소
- `address_type` (optional): 주소 타입
  - `ROAD`: 도로명 주소로만 검색 (빠름)
  - `PARCEL`: 지번 주소로만 검색
  - 생략 시: 자동으로 ROAD → PARCEL 순서로 폴백

**Success Response (200):**
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
        "zipcode": "04524",
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
    "processing_time_ms": 123000000
}
```

**Error Response (404):**
```json
{
    "success": false,
    "provider": "none",
    "attempts": [
        {
            "provider": "vWorld",
            "success": false,
            "error": "NOT_FOUND: 검색 결과가 없습니다"
        },
        {
            "provider": "Kakao",
            "success": false,
            "error": "[UNAUTHORIZED] Invalid API key: API key is invalid or expired"
        }
    ],
    "processed_at": "2025-11-25T10:00:00.000000+09:00",
    "processing_time_ms": 250000000,
    "error": "all providers failed to geocode the address"
}
```

#### POST /api/v1/geocode/bulk
Convert multiple Korean addresses to coordinates (max 100).

**Request:**
```json
{
    "addresses": [
        "서울시 강남구",
        "서울시 서초구",
        "경기도 성남시 분당구"
    ]
}
```

**Response (200):**
```json
{
    "results": [
        {
            "success": true,
            "address": {
                "input": "서울시 강남구",
                "formatted": "서울특별시 강남구"
            },
            "coordinates": {
                "lat": 37.517331,
                "lng": 127.047374
            },
            "provider": "vWorld",
            "processed_at": "2025-11-20T17:45:02.687159+09:00",
            "processing_time_ms": 32
        },
        {
            "success": false,
            "provider": "Kakao",
            "processed_at": "2025-11-20T17:45:02.679834+09:00",
            "processing_time_ms": 89,
            "error": "주소를 찾을 수 없습니다"
        }
    ],
    "summary": {
        "total": 3,
        "success": 2,
        "failed": 1
    },
    "processing_time_ms": 95
}
```

## Error Codes

- `400 Bad Request`: Invalid request format or parameters
- `404 Not Found`: Address not found (for single geocoding)
- `500 Internal Server Error`: Server error

## Request Headers

### Required Headers
- `Content-Type: application/json`

### Optional Headers
- `X-Request-ID`: Custom request ID for tracking (will be generated if not provided)

## Response Headers
- `X-Request-ID`: Request tracking ID
- `Access-Control-Allow-Origin`: CORS support

## Rate Limits
- vWorld: 25,000 requests/day
- Kakao: 300,000 requests/day

## Notes
1. All addresses must contain Korean characters
2. Coordinates are returned with 6 decimal places precision
3. The service automatically falls back from vWorld to Kakao if needed
4. Bulk requests are processed concurrently (max 10 concurrent)