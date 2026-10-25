package checkout

import (
	"fmt"

	"carf-chorf.com/app/pkg/money"
)

// CreditPackage is a purchasable prepaid credit bundle.
type CreditPackage struct {
	Credits     uint
	AmountCents uint
	Name        string
	Description string
	Badge       string
}

// Catalog is the server-side source of truth for package amounts. Update the
// placeholder prices here when the final pricing is available.
var Catalog = []CreditPackage{
	{
		Credits:     1,
		AmountCents: 1900,
		Name:        "1 Report Credit",
		Description: "One CARFAX report credit.",
	},
	{
		Credits:     2,
		AmountCents: 3400,
		Name:        "2 Report Credits",
		Description: "Two credits at a lower per-report price.",
		Badge:       "Popular",
	},
	{
		Credits:     4,
		AmountCents: 5900,
		Name:        "4 Report Credits",
		Description: "Four credits, our best per-report value.",
		Badge:       "Best value",
	},
}

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

// Find returns the package for the exact credit count.
func Find(credits uint) (CreditPackage, bool) {
	for _, pkg := range Catalog {
		if pkg.Credits == credits {
			return pkg, true
		}
	}
	return CreditPackage{}, false
}

// PackageViews returns packages with display-ready prices and savings.
func PackageViews() []PackageView {
	views := make([]PackageView, 0, len(Catalog))
	base := Catalog[0].AmountCents

	for _, pkg := range Catalog {
		perCredit := pkg.AmountCents / pkg.Credits
		savings := ""
		if pkg.Credits > 1 {
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
