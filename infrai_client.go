package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type InfraiClient struct {
	baseURL string
	key     string
	http    *http.Client
}

func NewInfraiClient(baseURL, key string) *InfraiClient {
	return &InfraiClient{baseURL: baseURL, key: key, http: &http.Client{Timeout: 10 * time.Second}}
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c *InfraiClient) post(path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.http.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("infrai response: %w", err)
		}
		if env.OK {
			if out != nil {
				return json.Unmarshal(env.Data, out)
			}
			return nil
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
			continue
		}
		return fmt.Errorf("infrai request rejected: %s", string(env.Error))
	}
	return fmt.Errorf("infrai request exhausted retries")
}

func (c *InfraiClient) SendPhoneCode(phone, purpose string) error {
	// infrai.auth.phone.send_code -> POST /v1/auth/phone/send_code
	return c.post("/auth/phone/send_code", map[string]string{"phone": phone, "purpose": purpose, "locale": "en-US"}, nil)
}

func (c *InfraiClient) VerifyPhone(phone, code string, login bool) (string, error) {
	// infrai.auth.phone.verify -> POST /v1/auth/phone/verify
	var out struct {
		SessionID string `json:"session_id"`
	}
	err := c.post("/auth/phone/verify", map[string]any{"phone": phone, "code": code, "login": login}, &out)
	return out.SessionID, err
}

func (c *InfraiClient) SendDeliveryCode(phone string) error {
	// infrai.sms.otp -> POST /v1/sms/otp
	return c.post("/sms/otp", map[string]string{"to": phone, "phone": phone}, nil)
}
