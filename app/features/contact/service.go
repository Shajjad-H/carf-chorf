package contact

import "carf-chorf.com/app/app/services/discord_service"

func send(req ContactRequest) error {
	return discord_service.SendContactNotification(req.Email, req.Phone, req.Message)
}
