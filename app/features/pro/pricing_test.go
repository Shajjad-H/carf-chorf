package pro

import (
	"errors"
	"testing"

	"carf-chorf.com/app/app/kernel"
)

func setPricing(t *testing.T) {
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
	kernel.Prices.Packages = []kernel.PricePackage{{Credits: 1, AmountCents: 4999}}
	t.Cleanup(func() {
		kernel.ProPricing.Tiers = prevTiers
		kernel.Prices.Packages = prevPackages
	})
}

func TestQuoteFor(t *testing.T) {
	setPricing(t)

	tests := []struct {
		reports   uint
		perReport uint
		total     uint
		savings   uint
		wantErr   error
	}{
		{reports: 9, wantErr: ErrReportsOutOfRange},
		{reports: 10, perReport: 1999, total: 19990, savings: 60},
		{reports: 24, perReport: 1999, total: 47976, savings: 60},
		{reports: 25, perReport: 1499, total: 37475, savings: 70},
		{reports: 37, perReport: 1499, total: 55463, savings: 70},
		{reports: 49, perReport: 1499, total: 73451, savings: 70},
		{reports: 50, perReport: 1299, total: 64950, savings: 74},
		{reports: 99, perReport: 1299, total: 128601, savings: 74},
		{reports: 100, perReport: 999, total: 99900, savings: 80},
		{reports: 249, perReport: 999, total: 248751, savings: 80},
		{reports: 250, perReport: 499, total: 124750, savings: 90},
		{reports: 500, perReport: 499, total: 249500, savings: 90},
		{reports: 501, wantErr: ErrReportsOutOfRange},
	}

	for _, tc := range tests {
		quote, err := QuoteFor(tc.reports)
		if tc.wantErr != nil {
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("QuoteFor(%d) error = %v, want %v", tc.reports, err, tc.wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("QuoteFor(%d) error = %v", tc.reports, err)
		}
		if quote.PerReportCents != tc.perReport || quote.TotalCents != tc.total || quote.SavingsPercent != tc.savings {
			t.Fatalf("QuoteFor(%d) = %+v, want per=%d total=%d savings=%d", tc.reports, quote, tc.perReport, tc.total, tc.savings)
		}
	}
}

func TestQuoteForNoTiers(t *testing.T) {
	prev := kernel.ProPricing.Tiers
	kernel.ProPricing.Tiers = nil
	t.Cleanup(func() { kernel.ProPricing.Tiers = prev })

	if _, err := QuoteFor(25); !errors.Is(err, ErrNoTiers) {
		t.Fatalf("QuoteFor() error = %v, want %v", err, ErrNoTiers)
	}
}
