package checkout

import (
	"fmt"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/money"
)

// PackageView is the formatted package data used by the template.
type PackageView struct {
	Credits     uint
	Price       string
	PerCredit   string
	Savings     string
	Name        string
	Description string
	Badge       string
}

// Find returns the configured package for the exact credit count.
func Find(credits uint) (kernel.PricePackage, bool) {
	for _, pkg := range kernel.Prices.Packages {
		if pkg.Credits == credits {
			return pkg, true
		}
	}
	return kernel.PricePackage{}, false
}

// PackageViews returns packages with display-ready prices and savings.
func PackageViews() []PackageView {
	packages := kernel.Prices.Packages
	views := make([]PackageView, 0, len(packages))
	if len(packages) == 0 {
		return views
	}

	base := uint(0)
	if basePkg, ok := Find(1); ok {
		base = basePkg.AmountCents
	}

	for _, pkg := range packages {
		if pkg.Credits == 0 {
			continue
		}

		perCredit := pkg.AmountCents / pkg.Credits
		savings := ""
		if pkg.Credits > 1 && base > 0 {
			fullPrice := base * pkg.Credits
			if pkg.AmountCents < fullPrice {
				percent := (fullPrice - pkg.AmountCents) * 100 / fullPrice
				savings = fmt.Sprintf("Save %d%%", percent)
			}
		}

		views = append(views, PackageView{
			Credits:     pkg.Credits,
			Price:       money.FormatCents(pkg.AmountCents),
			PerCredit:   money.FormatCents(perCredit),
			Savings:     savings,
			Name:        pkg.Name,
			Description: pkg.Description,
			Badge:       pkg.Badge,
		})
	}

	return views
}
