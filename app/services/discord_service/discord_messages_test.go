package discord_service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"carf-chorf.com/app/app/kernel"
)

func TestSendSOSNotification(t *testing.T) {
	var received string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		received = payload.Content
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	previous := kernel.Discord.SOSWebhookURL
	kernel.Discord.SOSWebhookURL = server.URL
	t.Cleanup(func() {
		kernel.Discord.SOSWebhookURL = previous
	})

	if err := SendSOSNotification("report failed"); err != nil {
		t.Fatalf("SendSOSNotification() error = %v", err)
	}
	if received != "report failed" {
		t.Fatalf("content = %q", received)
	}
}

func TestSendSOSNotificationRequiresWebhook(t *testing.T) {
	previous := kernel.Discord.SOSWebhookURL
	kernel.Discord.SOSWebhookURL = ""
	t.Cleanup(func() {
		kernel.Discord.SOSWebhookURL = previous
	})

	if err := SendSOSNotification("report failed"); err == nil {
		t.Fatal("SendSOSNotification() unexpectedly succeeded")
	}
}
