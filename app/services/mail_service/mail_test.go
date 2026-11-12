package mail_service

import (
	"strings"
	"testing"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/renderer"
)

func TestEmailTemplates(t *testing.T) {
	previousDomain := kernel.Server.Domain
	kernel.Server.Domain = "https://example.test"
	t.Cleanup(func() {
		kernel.Server.Domain = previousDomain
	})

	renderer.InitTemplates("../../../templates")

	otpCtx := emailContext("user@example.com", "Your verification code")
	otpCtx["code"] = "ABCD1234"
	otpCtx["expires_minutes"] = 10

	otpHTML, err := renderer.Render("emails/otp.html", otpCtx)
	if err != nil {
		t.Fatalf("render otp html: %v", err)
	}
	if !strings.Contains(otpHTML, "ABCD1234") || !strings.Contains(otpHTML, "Your sign-in code") {
		t.Fatal("otp html is missing expected content")
	}
	if !strings.Contains(otpHTML, "https://example.test/public/logo-transparent.png") {
		t.Fatal("otp html is missing the absolute logo URL")
	}

	otpText, err := renderer.Render("emails/otp.txt", otpCtx)
	if err != nil {
		t.Fatalf("render otp text: %v", err)
	}
	if !strings.Contains(otpText, "ABCD1234") || !strings.Contains(otpText, "10 min") {
		t.Fatal("otp text is missing expected content")
	}

	reportCtx := emailContext("user@example.com", "Your vehicle report is ready")
	reportCtx["report_url"] = "https://example.test/reports/view?access_token=abc"
	reportCtx["vin"] = "1HGCM82633A004352"
	reportCtx["vehicle"] = "2003 Honda Accord"

	reportHTML, err := renderer.Render("emails/report.html", reportCtx)
	if err != nil {
		t.Fatalf("render report html: %v", err)
	}
	for _, want := range []string{"2003 Honda Accord", "1HGCM82633A004352", "https://example.test/reports/view?access_token=abc", "View my report"} {
		if !strings.Contains(reportHTML, want) {
			t.Fatalf("report html is missing %q", want)
		}
	}

	reportText, err := renderer.Render("emails/report.txt", reportCtx)
	if err != nil {
		t.Fatalf("render report text: %v", err)
	}
	if !strings.Contains(reportText, "1HGCM82633A004352") || !strings.Contains(reportText, "https://example.test/reports/view?access_token=abc") {
		t.Fatal("report text is missing expected content")
	}

	creditCtx := emailContext("user@example.com", "You still have unused report credits.")
	creditCtx["credits"] = 3
	creditCtx["dashboard_url"] = "https://example.test/dashboard"

	creditHTML, err := renderer.Render("emails/credit_reminder.html", creditCtx)
	if err != nil {
		t.Fatalf("render credit reminder html: %v", err)
	}
	if !strings.Contains(creditHTML, "3 CARFAX reports to use") || !strings.Contains(creditHTML, "https://example.test/dashboard") {
		t.Fatal("credit reminder html is missing expected content")
	}

	creditText, err := renderer.Render("emails/credit_reminder.txt", creditCtx)
	if err != nil {
		t.Fatalf("render credit reminder text: %v", err)
	}
	if !strings.Contains(creditText, "3 full CARFAX reports to use") || !strings.Contains(creditText, "https://example.test/dashboard") {
		t.Fatal("credit reminder text is missing expected content")
	}

	paymentCtx := emailContext("user@example.com", "Your payment is confirmed.")
	paymentCtx["credits"] = 3
	paymentCtx["amount"] = "$99.90"
	paymentCtx["transaction_ref"] = "550e8400-e29b-41d4-a716-446655440000"
	paymentCtx["purchased_at"] = "11/05/2026 3:04 PM"
	paymentCtx["login_url"] = "https://example.test/login?next=%2Fdashboard"

	paymentHTML, err := renderer.Render("emails/payment_confirmation.html", paymentCtx)
	if err != nil {
		t.Fatalf("render payment confirmation html: %v", err)
	}
	for _, want := range []string{"+3 credits", "$99.90", "550e8400-e29b-41d4-a716-446655440000", "11/05/2026 3:04 PM", "https://example.test/login?next=%2Fdashboard"} {
		if !strings.Contains(paymentHTML, want) {
			t.Fatalf("payment confirmation html is missing %q", want)
		}
	}

	paymentText, err := renderer.Render("emails/payment_confirmation.txt", paymentCtx)
	if err != nil {
		t.Fatalf("render payment confirmation text: %v", err)
	}
	if !strings.Contains(paymentText, "3 credits") || !strings.Contains(paymentText, "$99.90") || !strings.Contains(paymentText, "550e8400-e29b-41d4-a716-446655440000") {
		t.Fatal("payment confirmation text is missing expected content")
	}
}
