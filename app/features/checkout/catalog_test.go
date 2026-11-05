package checkout

import (
	"testing"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/kernel"
)

func setTestPrices(t *testing.T) {
	t.Helper()

	previous := kernel.Prices.Packages
	kernel.Prices.Packages = []kernel.PricePackage{
		{Credits: 3, AmountCents: 9990, Name: "3 Report Credits", Description: "Best for comparing.", Badge: "We recommend"},
		{Credits: 1, AmountCents: 4999, Name: "1 Report Credit", Description: "Know everything."},
		{Credits: 5, AmountCents: 13499, Name: "5 Report Credits", Description: "Compare before you choose."},
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
		{1, 4999},
		{3, 9990},
		{5, 13499},
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

	if _, ok := Find(2); ok {
		t.Fatal("Find(2) unexpectedly succeeded")
	}
}

func TestPackageViews(t *testing.T) {
	setTestPrices(t)

	views := PackageViews()
	if len(views) != 3 {
		t.Fatalf("PackageViews() length = %d", len(views))
	}
	if views[0].Credits != 3 || views[0].Price != "$99.90" || views[0].PerCredit != "$33.30" || views[0].Savings != "Save 33%" || views[0].SavingsAmount != "$50.07" {
		t.Fatalf("recommended package = %+v", views[0])
	}
	if views[1].Credits != 1 || views[1].Price != "$49.99" || views[1].PerCredit != "$49.99" || views[1].Savings != "" || views[1].SavingsAmount != "" {
		t.Fatalf("single-credit package = %+v", views[1])
	}
	if views[2].Credits != 5 || views[2].Price != "$134.99" || views[2].PerCredit != "$27.00" || views[2].Savings != "Save 46%" || views[2].SavingsAmount != "$114.96" {
		t.Fatalf("five-credit package = %+v", views[2])
	}
}

func TestPackageViewsUsesSingleCreditBase(t *testing.T) {
	previous := kernel.Prices.Packages
	kernel.Prices.Packages = []kernel.PricePackage{
		{Credits: 5, AmountCents: 13499, Name: "5 Report Credits"},
		{Credits: 3, AmountCents: 9990, Name: "3 Report Credits"},
		{Credits: 1, AmountCents: 4999, Name: "1 Report Credit"},
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
	if views[2].Savings != "" {
		t.Fatal("single-credit package should not show savings")
	}
}

func TestSelectedCheckoutCreditsAllowsGuestMultiCredit(t *testing.T) {
	record := &domaine.CacheRecord{}

	if got := selectedCheckoutCredits(3, record, false); got != 3 {
		t.Fatalf("guest with vehicle selected %d credits, want 3", got)
	}
	if got := selectedCheckoutCredits(5, record, false); got != 5 {
		t.Fatalf("guest with vehicle selected %d credits, want 5", got)
	}
	if got := selectedCheckoutCredits(3, nil, false); got != 0 {
		t.Fatalf("guest without vehicle selected %d credits, want 0", got)
	}
	if got := selectedCheckoutCredits(5, nil, true); got != 5 {
		t.Fatalf("authenticated user selected %d credits, want 5", got)
	}
}
