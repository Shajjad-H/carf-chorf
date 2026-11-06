// Package pages renders the public pricing and legal pages.
package pages

import (
	"net/http"

	"carf-chorf.com/app/app/features/checkout"
	"carf-chorf.com/app/pkg/renderer"
	"github.com/flosch/pongo2/v7"
)

// Pricing renders the public credit pricing page.
func Pricing(w http.ResponseWriter, r *http.Request) {
	renderer.Pongo(w, r, "pages/pricing.html", http.StatusOK, pongo2.Context{
		"packages": checkout.PackageViews(),
	})
}

// Terms renders the Terms and Conditions page.
func Terms(w http.ResponseWriter, r *http.Request) {
	renderer.Pongo(w, r, "pages/terms.html", http.StatusOK, nil)
}

// CGV renders the Terms of Sale (CGV) page.
func CGV(w http.ResponseWriter, r *http.Request) {
	renderer.Pongo(w, r, "pages/cgv.html", http.StatusOK, nil)
}

// SampleReport renders an illustrative report preview.
func SampleReport(w http.ResponseWriter, r *http.Request) {
	renderer.Pongo(w, r, "pages/sample_report.html", http.StatusOK, pongo2.Context{
		"packages": checkout.PackageViews(),
	})
}

// Privacy renders the privacy policy.
func Privacy(w http.ResponseWriter, r *http.Request) {
	renderer.Pongo(w, r, "pages/privacy.html", http.StatusOK, nil)
}

// LegalNotice renders the legal notice.
func LegalNotice(w http.ResponseWriter, r *http.Request) {
	renderer.Pongo(w, r, "pages/legal_notice.html", http.StatusOK, nil)
}
