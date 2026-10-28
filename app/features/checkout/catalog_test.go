package checkout

import (
	"testing"

	"carf-chorf.com/app/app/kernel"
)

func setTestPrices(t *testing.T) {
	t.Helper()

	previous := kernel.Prices.Packages
	kernel.Prices.Packages = []kernel.PricePackage{
		{Credits: 1, AmountCents: 1900, Name: "1 Report Credit", Description: "One CARFAX report credit."},
		{Credits: 2, AmountCents: 3400, Name: "2 Report Credits", Description: "Two credits at a lower per-report price.", Badge: "Popular"},
		{Credits: 4, AmountCents: 5900, Name: "4 Report Credits", Description: "Four credits, our best per-report value.", Badge: "Best value"},
	}
	t.Cleanup(func() {
		kernel.Prices.Packages = previous
	})
}

func TestFind(t *testing.T) {
	setTestPrices(t)

	tests := []struct {
		credits uint
		amount  uint
	}{
		{1, 1900},
		{2, 3400},
		{4, 5900},
	}

	for _, tc := range tests {
		pkg, ok := Find(tc.credits)
		if !ok {
			t.Fatalf("Find(%d) not found", tc.credits)
		}
		if pkg.AmountCents != tc.amount {
			t.Fatalf("Find(%d).AmountCents = %d, want %d", tc.credits, pkg.AmountCents, tc.amount)
		}
	}

	if _, ok := Find(3); ok {
		t.Fatal("Find(3) unexpectedly succeeded")
	}
}

func TestPackageViews(t *testing.T) {
	setTestPrices(t)

	views := PackageViews()
	if len(views) != 3 {
		t.Fatalf("PackageViews() length = %d", len(views))
	}
	if views[0].Price != "$19.00" {
		t.Fatalf("first price = %q", views[0].Price)
	}
	if views[1].Savings == "" || views[2].Savings == "" {
		t.Fatal("multi-credit packages should show savings")
	}
}

func TestPackageViewsUsesSingleCreditBase(t *testing.T) {
	previous := kernel.Prices.Packages
	kernel.Prices.Packages = []kernel.PricePackage{
		{Credits: 4, AmountCents: 5900, Name: "4 Report Credits"},
		{Credits: 2, AmountCents: 3400, Name: "2 Report Credits"},
		{Credits: 1, AmountCents: 1900, Name: "1 Report Credit"},
	}
	t.Cleanup(func() {
		kernel.Prices.Packages = previous
	})

	views := PackageViews()
	if len(views) != 3 {
		t.Fatalf("PackageViews() length = %d", len(views))
	}
	if views[0].Savings == "" || views[1].Savings == "" {
		t.Fatal("savings should use the 1-credit package as the base regardless of order")
	}
}
