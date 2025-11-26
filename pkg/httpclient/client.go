package httpclient

import (
	"net"
	"net/http"
	"time"
)

// Client 최적화된 HTTP 클라이언트
type Client struct {
	*http.Client
}

// NewClient HTTP 클라이언트 생성
// Connection Pooling과 타임아웃 최적화 적용
func NewClient(timeout time.Duration) *Client {
	return &Client{
		Client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,              // 전체 유휴 연결 최대 수
				MaxIdleConnsPerHost:   10,               // 호스트당 유휴 연결 수
				IdleConnTimeout:       90 * time.Second, // 유휴 연결 타임아웃
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				DisableCompression:    false,
				ForceAttemptHTTP2:     true, // HTTP/2 활성화
			},
		},
	}
}

// DefaultClient 기본 설정의 HTTP 클라이언트
func DefaultClient() *Client {
	return NewClient(30 * time.Second)
}