// Package checkout implements credit package selection and transaction creation.
package checkout

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	auth_service "carf-chorf.com/app/app/services/auth"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
)

// Index renders the checkout page for a vehicle lookup or for a logged-in user
// buying credits directly.
func Index(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))

	record, err := loadCheckoutVehicle(token)
	if err != nil {
		if !errors.Is(err, ErrVehicleNotFound) {
			simple_logger.Errorf("checkout: find vehicle: %v", err)
		}
		redirectWithFlash(w, r, "/", "Vehicle record not found. Please run a new lookup.")
		return
	}

	rawCredits := strings.TrimSpace(r.URL.Query().Get("credits"))
	requestedCredits := parseCreditCount(rawCredits)
	if rawCredits != "" && requestedCredits == 0 {
		redirectWithFlash(w, r, "/pricing", "Choose a valid package.")
		return
	}
	if requestedCredits != 0 {
		if _, ok := PurchaseAmountCents(requestedCredits); !ok {
			if requestedCredits < ProMinCredits {
				redirectWithFlash(w, r, "/pricing", "Pro pricing starts at 10 reports. Choose a regular pack.")
			} else {
				redirectWithFlash(w, r, "/contact", "For more than 500 reports, contact us for a custom quote.")
			}
			return
		}
	}

	user, authenticated, err := loadCheckoutUser(r)
	if err != nil {
		if authenticated {
			_ = auth_service.ClearCurrent(w, r)
		}
		next := buildCheckoutURL(token, requestedCredits)
		redirectWithFlash(w, r, "/login?next="+url.QueryEscape(next), "Please log in again.")
		return
	}

	selectedCredits := selectedCheckoutCredits(requestedCredits, record, authenticated)
	ctx := checkoutPageContext(token, requestedCredits, selectedCredits, record, user)
	renderer.Pongo(w, r, "checkout/index.html", http.StatusOK, ctx)
}

// Create validates the purchase, creates the transaction, and sends the buyer
// to the payment page.
func Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form submission", http.StatusBadRequest)
		return
	}

	req := parsePurchaseRequest(r)
	if _, ok := PurchaseAmountCents(req.Credits); !ok {
		redirectWithFlash(w, r, buildCheckoutURL(req.Token, 0), invalidCreditsMessage(req.Credits))
		return
	}

	state, authenticated := auth_service.Current(r)

	var current *auth_service.State
	if authenticated {
		current = &state
	}

	transaction, err := CreatePurchase(current, req)
	if err != nil {
		handlePurchaseError(w, r, req, authenticated, err)
		return
	}

	http.Redirect(w, r, "/payment?token="+url.QueryEscape(transaction.Token), http.StatusSeeOther)
}
