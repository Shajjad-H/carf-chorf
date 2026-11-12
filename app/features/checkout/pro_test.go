package checkout

import (
	"testing"

	"carf-chorf.com/app/app/kernel"
)

func setProPricing(t *testing.T) {
	t.Helper()
	prevTiers := kernel.ProPricing.Tiers
	prevPackages := kernel.Prices.Packages
	kernel.ProPricing.Tiers = []kernel.ProPriceTier{
		{MinReports: 10, PerReportCents: 1999},
		{MinReports: 25, PerReportCents: 1499},
		{MinReports: 50, PerReportCents: 1299},
		{MinReports: 100, PerReportCents: 999},
		{MinReports: 250, PerReportCents: 499},
	}
	kernel.Prices.Packages = []kernel.PricePackage{
		{Credits: 1, AmountCents: 4999},
		{Credits: 3, AmountCents: 9990},
		{Credits: 5, AmountCents: 13499},
	}
	t.Cleanup(func() {
		kernel.ProPricing.Tiers = prevTiers
		kernel.Prices.Packages = prevPackages
	})
}

func TestProTotalCents(t *testing.T) {
	setProPricing(t)

	tests := []struct {
		credits uint
		total   uint
		ok      bool
	}{
		{credits: 9, ok: false},
		{credits: 10, total: 19990, ok: true},
		{credits: 24, total: 47976, ok: true},
		{credits: 25, total: 37475, ok: true},
		{credits: 37, total: 55463, ok: true},
		{credits: 49, total: 73451, ok: true},
		{credits: 50, total: 64950, ok: true},
		{credits: 100, total: 99900, ok: true},
		{credits: 249, total: 248751, ok: true},
		{credits: 250, total: 124750, ok: true},
		{credits: 500, total: 249500, ok: true},
		{credits: 501, ok: false},
	}

	for _, tc := range tests {
		total, ok := ProTotalCents(tc.credits)
		if ok != tc.ok || total != tc.total {
			t.Fatalf("ProTotalCents(%d) = (%d, %v), want (%d, %v)", tc.credits, total, ok, tc.total, tc.ok)
		}
	}
}

func TestCheckoutPageContextExposesProLine(t *testing.T) {
	setProPricing(t)

	ctx := checkoutPageContext("", 37, 37, nil, nil)
	if ctx["is_pro"] != true {
		t.Fatal("checkout context should mark a pro quantity")
	}
	if ctx["pro_credits"] != uint(37) || ctx["pro_per_report"] != "$14.99" || ctx["pro_total"] != "$554.63" {
		t.Fatalf("unexpected pro context: %#v", ctx)
	}
}

func TestPurchaseAmountCentsKeepsPackPricing(t *testing.T) {
	setProPricing(t)

	tests := []struct {
		credits uint
		total   uint
	}{
		{credits: 1, total: 4999},
		{credits: 3, total: 9990},
		{credits: 5, total: 13499},
		{credits: 37, total: 55463},
	}

	for _, tc := range tests {
		total, ok := PurchaseAmountCents(tc.credits)
		if !ok || total != tc.total {
			t.Fatalf("PurchaseAmountCents(%d) = (%d, %v), want (%d, true)", tc.credits, total, ok, tc.total)
		}
	}
}
