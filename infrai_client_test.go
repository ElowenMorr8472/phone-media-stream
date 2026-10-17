package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCapabilityRequests(t *testing.T) {
	tests := []struct {
		name, path string
		fields     map[string]any
		call       func(*InfraiClient) error
	}{
		{"send_code", "/v1/auth/phone/send_code", map[string]any{"phone": "+15551234567"}, func(c *InfraiClient) error {
			return c.SendPhoneCode("+15551234567", "signup")
		}},
		{"verify", "/v1/auth/phone/verify", map[string]any{"phone": "+15551234567", "code": "123456"}, func(c *InfraiClient) error {
			_, err := c.VerifyPhone("+15551234567", "123456", true)
			return err
		}},
		{"otp", "/v1/sms/otp", map[string]any{"to": "+15551234567"}, func(c *InfraiClient) error {
			return c.SendDeliveryCode("+15551234567")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.path {
					t.Errorf("path = %q, want %q", r.URL.Path, tt.path)
				}
				defer r.Body.Close()
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
					return
				}
				var fields map[string]any
				if err := json.Unmarshal(body, &fields); err != nil {
					t.Error(err)
					return
				}
				for key, want := range tt.fields {
					if fields[key] != want {
						t.Errorf("%s = %v, want %v", key, fields[key], want)
					}
				}
				_, _ = w.Write([]byte(`{"ok":true,"data":{"session_id":"test-session"}}`))
			}))
			defer server.Close()
			if err := tt.call(NewInfraiClient(server.URL+"/v1", "test-key")); err != nil {
				t.Fatal(err)
			}
		})
	}
}
