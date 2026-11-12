// Package pro implements the volume pricing calculator and purchase flow.
package pro

import (
	"errors"
	"fmt"
	"math"

	"carf-chorf.com/app/app/features/checkout"
	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/money"
)

const (
	// MinReports is the smallest order accepted by the pro calculator.
	MinReports uint = checkout.ProMinCredits
	// MaxReports keeps the calculator within the published volume bands.
	MaxReports uint = checkout.ProMaxCredits
)

var (
	ErrReportsOutOfRange = errors.New("reports out of range")
	ErrNoTiers           = errors.New("no pro pricing tiers configured")
)

// Quote is the server-side price for a volume order.
type Quote struct {
	Reports           uint
	PerReportCents    uint
	TotalCents        uint
	SingleReportCents uint
	SavingsPercent    uint
	PerReport         string
	Total             string
	SingleReport      string
}

// SingleReportCents returns the configured 1-credit package price.
func SingleReportCents() uint {
	for _, pkg := range kernel.Prices.Packages {
		if pkg.Credits == 1 {
			return pkg.AmountCents
		}
	}
	return 0
}

// QuoteFor returns the volume price for reports.
func QuoteFor(reports uint) (Quote, error) {
	if reports < MinReports || reports > MaxReports {
		return Quote{}, fmt.Errorf("%w: choose between %d and %d reports", ErrReportsOutOfRange, MinReports, MaxReports)
	}
	if len(kernel.ProPricing.Tiers) == 0 {
		return Quote{}, ErrNoTiers
	}

	var tier *kernel.ProPriceTier
	for i := range kernel.ProPricing.Tiers {
		candidate := &kernel.ProPricing.Tiers[i]
		if reports >= candidate.MinReports {
			tier = candidate
		}
	}
	if tier == nil {
		return Quote{}, ErrReportsOutOfRange
	}

	single := SingleReportCents()
	savings := uint(0)
	if single > 0 && tier.PerReportCents < single {
		savings = uint(math.Round(float64(single-tier.PerReportCents) * 100 / float64(single)))
	}

	total := tier.PerReportCents * reports
	return Quote{
		Reports:           reports,
		PerReportCents:    tier.PerReportCents,
		TotalCents:        total,
		SingleReportCents: single,
		SavingsPercent:    savings,
		PerReport:         money.FormatCents(tier.PerReportCents),
		Total:             money.FormatCents(total),
		SingleReport:      money.FormatCents(single),
	}, nil
}
