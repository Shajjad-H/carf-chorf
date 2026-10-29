package mail_service

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/renderer"
	"github.com/flosch/pongo2/v7"
)

// SendOTP sends a login verification code.
func SendOTP(to, code string) error {
	ctx := emailContext(to, "Your verification code")
	ctx["code"] = code
	ctx["expires_minutes"] = 10

	html, err := renderer.Render("emails/otp.html", ctx)
	if err != nil {
		return fmt.Errorf("mail_service: render otp html: %w", err)
	}
	text, err := renderer.Render("emails/otp.txt", ctx)
	if err != nil {
		return fmt.Errorf("mail_service: render otp text: %w", err)
	}

	return send(to, "Your verification code", text, html)
}

// SendReport sends the link to a purchased vehicle report.
func SendReport(to, reportURL, vin, vehicle string) error {
	ctx := emailContext(to, "Your vehicle report is ready to view.")
	ctx["report_url"] = reportURL
	ctx["vin"] = vin
	ctx["vehicle"] = vehicle

	html, err := renderer.Render("emails/report.html", ctx)
	if err != nil {
		return fmt.Errorf("mail_service: render report html: %w", err)
	}
	text, err := renderer.Render("emails/report.txt", ctx)
	if err != nil {
		return fmt.Errorf("mail_service: render report text: %w", err)
	}

	return send(to, "Your vehicle report is ready", text, html)
}

func emailContext(to, preheader string) pongo2.Context {
	domain := strings.TrimRight(kernel.Server.Domain, "/")
	return pongo2.Context{
		"brand_name":  "Vehicle Report",
		"domain":      domain,
		"logo_url":    domain + "/public/logo-original-compressed.png",
		"contact_url": domain + "/contact",
		"pricing_url": domain + "/pricing",
		"terms_url":   domain + "/terms",
		"year":        time.Now().Year(),
		"to":          to,
		"preheader":   preheader,
	}
}

func send(to, subject, text, html string) error {
	cfg := kernel.Mail

	from := cfg.From
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.From)
	}

	const boundary = "carf-chorf-message"

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
		return fmt.Errorf("mail_service: send: %w", err)
	}
	return nil
}
