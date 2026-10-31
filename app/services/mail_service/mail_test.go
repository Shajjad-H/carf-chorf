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
	if !strings.Contains(otpHTML, "ABCD1234") || !strings.Contains(otpHTML, "Verify your email") {
		t.Fatal("otp html is missing expected content")
	}
	if !strings.Contains(otpHTML, "https://example.test/public/logo-original-compressed.png") {
		t.Fatal("otp html is missing the absolute logo URL")
	}

	otpText, err := renderer.Render("emails/otp.txt", otpCtx)
	if err != nil {
		t.Fatalf("render otp text: %v", err)
	}
	if !strings.Contains(otpText, "ABCD1234") || !strings.Contains(otpText, "10 minutes") {
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
	for _, want := range []string{"2003 Honda Accord", "1HGCM82633A004352", "https://example.test/reports/view?access_token=abc", "View your report"} {
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
	if !strings.Contains(creditHTML, "3 unused credits") || !strings.Contains(creditHTML, "https://example.test/dashboard") {
		t.Fatal("credit reminder html is missing expected content")
	}

	creditText, err := renderer.Render("emails/credit_reminder.txt", creditCtx)
	if err != nil {
		t.Fatalf("render credit reminder text: %v", err)
	}
	if !strings.Contains(creditText, "3 unused credits") || !strings.Contains(creditText, "https://example.test/dashboard") {
		t.Fatal("credit reminder text is missing expected content")
	}
}
