package discord_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/http_client"
)

const maxContentLength = 2000

func SendContactNotification(email, phone, message string) error {
	content := fmt.Sprintf(
		"📬 **New contact message**\n**Email:** %s\n**Phone:** %s\n**Message:**\n%s",
		email, phone, message,
	)
	return sendNotification(kernel.Discord.ContactWebhookURL, content)
}

func sendNotification(webhookURL, content string) error {
	if webhookURL == "" {
		return fmt.Errorf("discord_service: webhook not configured")
	}

	payload := map[string]any{
		"content":          truncate(content, maxContentLength),
		"allowed_mentions": map[string]any{"parse": []string{}},
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http_client.Shared.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("discord API error: received status code %d with message %s", resp.StatusCode, string(body))
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-3]) + "..."
}
