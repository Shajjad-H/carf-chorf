package checkout

import "testing"

func TestFind(t *testing.T) {
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
