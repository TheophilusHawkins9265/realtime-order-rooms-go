package orderchat

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublishRetriesRateLimitWithSameIdempotencyKey(t *testing.T) {
	attempts := 0
	var keys []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		keys = append(keys, r.Header.Get("Idempotency-Key"))
		if r.Method != http.MethodPost || r.URL.Path != "/v1/realtime/publish" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if attempts == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"ok":true,"data":{"published":true},"error":null,"metadata":{}}`)
	}))
	defer server.Close()

	client := NewRealtimeClient("test-key")
	client.baseURL = server.URL
	client.http = server.Client()
	client.sleep = func(context.Context, time.Duration) error { return nil }
	err := client.Publish(context.Background(), "order:ord_2048", "order.updated", map[string]string{"stage": "fulfillment.packing"}, "acct_7", "evt_91")
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 2 || len(keys) != 2 || keys[0] != "evt_91" || keys[1] != "evt_91" {
		t.Fatalf("attempts = %d, keys = %v", attempts, keys)
	}
}
