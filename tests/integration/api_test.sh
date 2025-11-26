#!/bin/bash

# Geocoding API Integration Test Script
# 이 스크립트는 Geocoding API 서버의 모든 엔드포인트를 테스트합니다.

set -e

# 색상 정의
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 서버 URL 설정
BASE_URL="${BASE_URL:-http://localhost:8080}"

# 결과 카운터
PASSED=0
FAILED=0

# 테스트 함수
test_endpoint() {
    local test_name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected_status=$5

    echo -e "\n${YELLOW}Testing: ${test_name}${NC}"
    echo "  Method: $method"
    echo "  Endpoint: $endpoint"

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "${BASE_URL}${endpoint}")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "${BASE_URL}${endpoint}")
    fi

    # 마지막 줄이 HTTP 상태 코드
    status_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')

    echo "  Status Code: $status_code (expected: $expected_status)"

    if [ "$status_code" = "$expected_status" ]; then
        echo -e "  ${GREEN}✓ PASSED${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "  ${RED}✗ FAILED${NC}"
        FAILED=$((FAILED + 1))
    fi

    # JSON 응답이면 pretty print
    if echo "$body" | jq . >/dev/null 2>&1; then
        echo "  Response:"
        echo "$body" | jq . | sed 's/^/    /'
    else
        echo "  Response: $body"
    fi
}

echo "========================================="
echo "Geocoding API Integration Test"
echo "Base URL: $BASE_URL"
echo "========================================="

# 1. Health Check 엔드포인트 테스트
echo -e "\n${YELLOW}=== Health Check Endpoints ===${NC}"

test_endpoint \
    "Ping Test" \
    "GET" \
    "/ping" \
    "" \
    "200"

test_endpoint \
    "Health Check" \
    "GET" \
    "/health" \
    "" \
    "200"

test_endpoint \
    "Readiness Check" \
    "GET" \
    "/ready" \
    "" \
    "200"

# 2. 단건 지오코딩 테스트
echo -e "\n${YELLOW}=== Single Geocoding Tests ===${NC}"

test_endpoint \
    "Valid Address - Seoul" \
    "POST" \
    "/api/v1/geocode" \
    '{"address": "서울특별시 강남구 테헤란로 152"}' \
    "200"

test_endpoint \
    "Valid Address - Pangyo" \
    "POST" \
    "/api/v1/geocode" \
    '{"address": "경기도 성남시 분당구 판교역로 166"}' \
    "200"

test_endpoint \
    "Empty Address" \
    "POST" \
    "/api/v1/geocode" \
    '{"address": ""}' \
    "400"

test_endpoint \
    "Invalid Short Address" \
    "POST" \
    "/api/v1/geocode" \
    '{"address": "123"}' \
    "404"

test_endpoint \
    "Missing Address Field" \
    "POST" \
    "/api/v1/geocode" \
    '{}' \
    "400"

test_endpoint \
    "Invalid JSON" \
    "POST" \
    "/api/v1/geocode" \
    'invalid json' \
    "400"

# 3. 대량 지오코딩 테스트
echo -e "\n${YELLOW}=== Bulk Geocoding Tests ===${NC}"

test_endpoint \
    "Bulk - 2 Valid Addresses" \
    "POST" \
    "/api/v1/geocode/bulk" \
    '{"addresses": ["서울특별시 강남구 테헤란로 152", "경기도 성남시 분당구 판교역로 166"]}' \
    "200"

test_endpoint \
    "Bulk - Single Address" \
    "POST" \
    "/api/v1/geocode/bulk" \
    '{"addresses": ["서울특별시 강남구 테헤란로 152"]}' \
    "200"

test_endpoint \
    "Bulk - Empty Array" \
    "POST" \
    "/api/v1/geocode/bulk" \
    '{"addresses": []}' \
    "400"

test_endpoint \
    "Bulk - Missing Addresses Field" \
    "POST" \
    "/api/v1/geocode/bulk" \
    '{}' \
    "400"

# 4. 에러 처리 테스트
echo -e "\n${YELLOW}=== Error Handling Tests ===${NC}"

test_endpoint \
    "404 - Invalid Endpoint" \
    "GET" \
    "/api/v1/invalid" \
    "" \
    "404"

test_endpoint \
    "405 - Wrong Method" \
    "GET" \
    "/api/v1/geocode" \
    "" \
    "404"

# 5. 대량 주소 테스트 (최대 개수)
echo -e "\n${YELLOW}=== Bulk Limit Tests ===${NC}"

# 10개 주소 생성
addresses_10='{"addresses": ['
for i in {1..10}; do
    addresses_10+="\"서울특별시 강남구 테헤란로 ${i}번지\""
    if [ $i -lt 10 ]; then
        addresses_10+=", "
    fi
done
addresses_10+=']}'

test_endpoint \
    "Bulk - 10 Addresses" \
    "POST" \
    "/api/v1/geocode/bulk" \
    "$addresses_10" \
    "200"

# 결과 요약
echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo -e "Total Tests: $((PASSED + FAILED))"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo "========================================="

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
