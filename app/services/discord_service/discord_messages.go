package discord_service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"carf-chorf.com/app/pkg/http_client"
)

// SendContactNotification envoie le message d'un visiteur du formulaire de
// contact dans le salon Discord configuré.
// func SendContactNotification(nom, email, sujet, message string) error {
// 	content := fmt.Sprintf(
// 		"📩 **Nouveau message de contact**\n```Nom : %s\nE-mail : %s\nSujet : %s\n\n%s```",
// 		nom, email, sujet, message,
// 	)

// 	return sendNotification(configs.Discord.ContactWebhookURL, content)
// }

func sendNotification(webhookURL, content string) error {
	if webhookURL == "" {
		return fmt.Errorf("discord_service: webhook non configuré")
	}

	payload := map[string]string{"content": content}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
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
