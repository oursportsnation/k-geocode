package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"short timeout", 1 * time.Second},
		{"medium timeout", 5 * time.Second},
		{"long timeout", 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.timeout)

			require.NotNil(t, client)
			assert.NotNil(t, client.Client)
			assert.Equal(t, tt.timeout, client.Client.Timeout)
		})
	}
}

func TestDefaultClient(t *testing.T) {
	client := DefaultClient()

	require.NotNil(t, client)
	assert.NotNil(t, client.Client)
	assert.Equal(t, 30*time.Second, client.Client.Timeout)
}

func TestClient_MakesHTTPRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)

	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_HandlesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(100 * time.Millisecond)

	_, err := client.Get(server.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}
