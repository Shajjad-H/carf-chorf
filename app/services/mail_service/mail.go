package mail_service

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"carf-chorf.com/app/app/kernel"
)

func SendOTP(to, code string) error {
	cfg := kernel.Mail

	from := cfg.From
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.From)
	}

	const subject = "Your verification code"
	text := fmt.Sprintf("Your verification code is %s.\n\nIt expires in 10 minutes.", code)
	html := fmt.Sprintf(`<!doctype html>
<html>
<body style="margin:0;padding:32px;background:#f1f5f9;font-family:Arial,Helvetica,sans-serif;color:#0f172a;">
  <div style="max-width:480px;margin:0 auto;background:#ffffff;border-radius:16px;padding:32px;border:1px solid #e2e8f0;">
    <h1 style="margin:0 0 8px;font-size:20px;">Your verification code</h1>
    <p style="margin:0 0 24px;color:#475569;font-size:14px;">Enter this code to continue. It expires in 10 minutes.</p>
    <div style="font-size:32px;font-weight:700;letter-spacing:8px;text-align:center;padding:16px;background:#eef2ff;border-radius:12px;color:#4338ca;">%s</div>
    <p style="margin:24px 0 0;color:#94a3b8;font-size:12px;">If you didn't request this code, you can safely ignore this email.</p>
  </div>
</body>
</html>`, code)

	const boundary = "carf-chorf-otp"

	var body strings.Builder
	fmt.Fprintf(&body, "From: %s\r\n", from)
	fmt.Fprintf(&body, "To: %s\r\n", to)
	fmt.Fprintf(&body, "Subject: %s\r\n", subject)
	fmt.Fprintf(&body, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&body, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
	fmt.Fprintf(&body, "--%s\r\n", boundary)
	fmt.Fprintf(&body, "Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	fmt.Fprintf(&body, "%s\r\n", text)
	fmt.Fprintf(&body, "--%s\r\n", boundary)
	fmt.Fprintf(&body, "Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	fmt.Fprintf(&body, "%s\r\n", html)
	fmt.Fprintf(&body, "--%s--\r\n", boundary)

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	if err := smtp.SendMail(addr, auth, cfg.From, []string{to}, []byte(body.String())); err != nil {
		return fmt.Errorf("mail_service: send otp: %w", err)
	}

	return nil
}
