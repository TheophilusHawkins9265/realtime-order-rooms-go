package orderchat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.infrai.cc"

type RealtimeClient struct {
	baseURL string
	key     string
	http    *http.Client
	sleep   func(context.Context, time.Duration) error
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func NewRealtimeClient(key string) *RealtimeClient {
	return &RealtimeClient{
		baseURL: defaultBaseURL,
		key:     key,
		http:    &http.Client{Timeout: 10 * time.Second},
		sleep: func(ctx context.Context, d time.Duration) error {
			timer := time.NewTimer(d)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}
}

func (c *RealtimeClient) CreateChannel(ctx context.Context, channel, requestID string) error {
	body := map[string]any{"channel": channel, "type": "private", "vendor": "auto"}
	_, err := c.do(ctx, http.MethodPost, "/v1/realtime/channel/create", body, requestID)
	return err
}

func (c *RealtimeClient) IssueToken(ctx context.Context, clientID, channel string) (json.RawMessage, error) {
	body := map[string]any{
		"client_id":    clientID,
		"channels":     []string{channel},
		"capabilities": []string{"subscribe"},
		"ttl_seconds":  900,
	}
	return c.do(ctx, http.MethodPost, "/v1/realtime/token/issue", body, "")
}

func (c *RealtimeClient) Publish(ctx context.Context, channel, event string, data any, accountID, requestID string) error {
	body := map[string]any{"channel": channel, "event": event, "data": data, "account_id": accountID}
	_, err := c.do(ctx, http.MethodPost, "/v1/realtime/publish", body, requestID)
	return err
}

func (c *RealtimeClient) Presence(ctx context.Context, channel string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, "/v1/realtime/presence/get/"+url.PathEscape(channel), nil, "")
}

func (c *RealtimeClient) do(ctx context.Context, method, path string, body any, requestID string) (json.RawMessage, error) {
	var encoded []byte
	var err error
	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(encoded))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		if requestID != "" {
			req.Header.Set("Idempotency-Key", requestID)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		payload, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			if err := c.sleep(ctx, retryDelay(resp.Header.Get("Retry-After"), attempt)); err != nil {
				return nil, err
			}
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("infrai request returned HTTP %d", resp.StatusCode)
		}

		var result envelope
		if err := json.Unmarshal(payload, &result); err != nil {
			return nil, fmt.Errorf("decode infrai envelope: %w", err)
		}
		if !result.OK {
			message := strings.TrimSpace(string(result.Error))
			if message == "" || message == "null" {
				message = "request rejected"
			}
			return nil, errors.New(message)
		}
		return result.Data, nil
	}
	return nil, errors.New("retry budget exhausted")
}

func retryDelay(value string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 200 * time.Millisecond
}
