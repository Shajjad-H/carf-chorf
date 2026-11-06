package pro

import (
	"fmt"
	"math"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/features/checkout"
	"carf-chorf.com/app/app/kernel"
	auth_service "carf-chorf.com/app/app/services/auth"
	flashes_service "carf-chorf.com/app/app/services/flashes"
	"carf-chorf.com/app/app/services/transaction_service"
	"carf-chorf.com/app/app/services/turnstile_service"
	"carf-chorf.com/app/app/services/user_service"
	"carf-chorf.com/app/pkg/money"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
)

// TierView is the display form of one volume tier.
type TierView struct {
	Index          int
	MinReports     uint
	MaxReports     uint
	RangeLabel     string
	PerReportCents uint
	PerReport      string
	PriceNumber    string
	TotalAtMin     string
	SavingsPercent uint
	BestValue      bool
	LowestPrice    bool
}

// Index renders the pro pricing calculator.
func Index(w http.ResponseWriter, r *http.Request) {
	rawInitial := r.URL.Query().Get("reports")
	if rawInitial == "" {
		rawInitial = r.URL.Query().Get("credits")
	}
	initial := initialReports(rawInitial)
	quote, err := QuoteFor(initial)
	if err != nil {
		initial = 25
		quote, _ = QuoteFor(initial)
	}

	renderer.Pongo(w, r, "pages/pro_pricing.html", http.StatusOK, pongo2.Context{
		"packages":           checkout.PackageViews(),
		"tiers":              TierViews(),
		"initial_reports":    initial,
		"quote":              quote,
		"single_price":       money.FormatCents(SingleReportCents()),
		"single_price_cents": SingleReportCents(),
	})
}

// CheckoutPage renders the confirmation step for a fixed pro quantity.
func CheckoutPage(w http.ResponseWriter, r *http.Request) {
	reports, ok := parseReports(r.URL.Query().Get("reports"))
	if !ok {
		flashes_service.AddFlash(w, r, "error", "Choose a quantity between 10 and 500 reports.")
		http.Redirect(w, r, "/pro", http.StatusSeeOther)
		return
	}
	quote, err := QuoteFor(reports)
	if err != nil {
		flashes_service.AddFlash(w, r, "error", "Choose a quantity between 10 and 500 reports.")
		http.Redirect(w, r, "/pro", http.StatusSeeOther)
		return
	}

	renderer.Pongo(w, r, "pro/checkout.html", http.StatusOK, pongo2.Context{
		"reports":      reports,
		"quote":        quote,
		"single_price": money.FormatCents(SingleReportCents()),
	})
}

// Create validates a pro order and starts the payment flow.
func Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form submission", http.StatusBadRequest)
		return
	}

	rawReports := strings.TrimSpace(r.PostFormValue("reports"))
	value, err := strconv.ParseUint(rawReports, 10, 32)
	if err != nil {
		proError(w, r, 0, "Choose a quantity between 10 and 500 reports.")
		return
	}
	reports := uint(value)

	quote, err := QuoteFor(reports)
	if err != nil {
		proError(w, r, 0, "Choose a quantity between 10 and 500 reports.")
		return
	}

	if r.PostFormValue("terms") != "on" {
		proError(w, r, reports, "Please accept the Terms of Service and Terms of Sale.")
		return
	}

	if err := turnstile_service.Verify(r.Context(), r.PostFormValue("cf-turnstile-response"), ""); err != nil {
		simple_logger.Warnf("pro: turnstile verification failed: %v", err)
		proError(w, r, reports, "Please complete the verification challenge and try again.")
		return
	}

	state, authenticated := auth_service.Current(r)
	var user *domaine.User
	if authenticated {
		user, err = user_service.FindByID(state.UserID)
	} else {
		email := normalizeEmail(r.PostFormValue("email"))
		confirm := normalizeEmail(r.PostFormValue("confirm_email"))
		if !validEmail(email) {
			proError(w, r, reports, "Enter a valid email address.")
			return
		}
		if email != confirm {
			proError(w, r, reports, "Email addresses do not match.")
			return
		}
		user, err = user_service.FindOrCreateByEmail(email)
	}
	if err != nil {
		simple_logger.Errorf("pro: find user: %v", err)
		proError(w, r, reports, "Something went wrong. Please try again.")
		return
	}
	if user == nil || !user.IsActive {
		proError(w, r, reports, "This account is not active.")
		return
	}

	transaction, err := transaction_service.Create(user.ID, nil, quote.Reports, quote.TotalCents)
	if err != nil {
		simple_logger.Errorf("pro: create transaction: %v", err)
		proError(w, r, reports, "Could not create the order. Please try again.")
		return
	}

	http.Redirect(w, r, "/payment?token="+url.QueryEscape(transaction.Token), http.StatusSeeOther)
}

func proError(w http.ResponseWriter, r *http.Request, reports uint, message string) {
	flashes_service.AddFlash(w, r, "error", message)
	target := "/pro"
	if reports >= MinReports && reports <= MaxReports {
		target = "/pro?reports=" + strconv.FormatUint(uint64(reports), 10)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

func initialReports(raw string) uint {
	value, ok := parseReports(raw)
	if !ok {
		return 25
	}
	return value
}

func parseReports(raw string) (uint, bool) {
	value, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 32)
	if err != nil || value < uint64(MinReports) || value > uint64(MaxReports) {
		return 0, false
	}
	return uint(value), true
}

// TierViews returns the configured volume tiers in display form.
func TierViews() []TierView {
	views := make([]TierView, 0, len(kernel.ProPricing.Tiers))
	single := SingleReportCents()
	for i, tier := range kernel.ProPricing.Tiers {
		max := MaxReports
		if i+1 < len(kernel.ProPricing.Tiers) {
			max = kernel.ProPricing.Tiers[i+1].MinReports - 1
		}
		label := fmt.Sprintf("%d–%d reports", tier.MinReports, max)
		view := TierView{
			Index:          i,
			MinReports:     tier.MinReports,
			MaxReports:     max,
			RangeLabel:     label,
			PerReportCents: tier.PerReportCents,
			PerReport:      money.FormatCents(tier.PerReportCents),
			PriceNumber:    money.Decimal(tier.PerReportCents),
			TotalAtMin:     money.FormatCents(tier.PerReportCents * tier.MinReports),
			BestValue:      tier.MinReports == 25,
			LowestPrice:    i == len(kernel.ProPricing.Tiers)-1,
		}
		if single > 0 && tier.PerReportCents < single {
			view.SavingsPercent = uint(math.Round(float64(single-tier.PerReportCents) * 100 / float64(single)))
		}
		views = append(views, view)
	}
	return views
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
