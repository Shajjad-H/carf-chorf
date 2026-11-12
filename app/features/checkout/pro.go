package checkout

import "carf-chorf.com/app/app/kernel"

const (
	// ProMinCredits is the smallest quantity eligible for pro pricing.
	ProMinCredits uint = 10
	// ProMaxCredits is the largest self-serve pro quantity.
	ProMaxCredits uint = 500
)

// ProTierFor returns the configured pro tier for a quantity. Quantities below
// the minimum or above the self-serve maximum are rejected.
func ProTierFor(credits uint) (kernel.ProPriceTier, bool) {
	if credits < ProMinCredits || credits > ProMaxCredits {
		return kernel.ProPriceTier{}, false
	}

	var tier *kernel.ProPriceTier
	for i := range kernel.ProPricing.Tiers {
		candidate := &kernel.ProPricing.Tiers[i]
		if credits >= candidate.MinReports {
			tier = candidate
		}
	}
	if tier == nil {
		return kernel.ProPriceTier{}, false
	}
	return *tier, true
}

// ProTotalCents returns the server-side total for a pro quantity.
func ProTotalCents(credits uint) (uint, bool) {
	tier, ok := ProTierFor(credits)
	if !ok {
		return 0, false
	}
	return credits * tier.PerReportCents, true
}

// PurchaseAmountCents returns the server-side amount for a public pack or a
// pro quantity.
func PurchaseAmountCents(credits uint) (uint, bool) {
	if pkg, ok := Find(credits); ok {
		return pkg.AmountCents, true
	}
	return ProTotalCents(credits)
}
