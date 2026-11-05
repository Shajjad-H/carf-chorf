package mail_service

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/money"
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

// SendCreditReminder reminds a user about unused report credits.
func SendCreditReminder(to string, credits uint) error {
	domain := strings.TrimRight(kernel.Server.Domain, "/")
	ctx := emailContext(to, "You still have unused report credits.")
	ctx["credits"] = credits
	ctx["dashboard_url"] = domain + "/dashboard"

	html, err := renderer.Render("emails/credit_reminder.html", ctx)
	if err != nil {
		return fmt.Errorf("mail_service: render credit reminder html: %w", err)
	}
	text, err := renderer.Render("emails/credit_reminder.txt", ctx)
	if err != nil {
		return fmt.Errorf("mail_service: render credit reminder text: %w", err)
	}

	return send(to, "You have unused report credits", text, html)
}

// SendPaymentConfirmation sends the receipt for a credit purchase.
func SendPaymentConfirmation(to string, credits, amountCents uint, transactionRef string, purchasedAt time.Time) error {
	domain := strings.TrimRight(kernel.Server.Domain, "/")
	ctx := emailContext(to, "Your payment is confirmed and your credits are ready.")
	ctx["credits"] = credits
	ctx["amount"] = money.FormatCents(amountCents)
	ctx["transaction_ref"] = transactionRef
	ctx["purchased_at"] = purchasedAt.Format("01/02/2006 3:04 PM")
	ctx["login_url"] = domain + "/login?next=%2Fdashboard"

	html, err := renderer.Render("emails/payment_confirmation.html", ctx)
	if err != nil {
		return fmt.Errorf("mail_service: render payment confirmation html: %w", err)
	}
	text, err := renderer.Render("emails/payment_confirmation.txt", ctx)
	if err != nil {
		return fmt.Errorf("mail_service: render payment confirmation text: %w", err)
	}

	return send(to, "Payment confirmed — Your credits are ready", text, html)
}

func emailContext(to, preheader string) pongo2.Context {
	domain := strings.TrimRight(kernel.Server.Domain, "/")
	return pongo2.Context{
		"brand_name":  "Vehicle Report",
		"domain":      domain,
		"logo_url":    domain + "/public/logo-transparent.png",
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
