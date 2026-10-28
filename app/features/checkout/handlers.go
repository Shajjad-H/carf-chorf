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

	requestedCredits := parseCredits(r.URL.Query().Get("credits"))
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
	if _, ok := Find(req.Credits); !ok {
		redirectWithFlash(w, r, buildCheckoutURL(req.Token, 0), "Please select a valid package.")
		return
	}

	state, authenticated := auth_service.Current(r)
	if !authenticated && req.Credits != 1 {
		redirectToLogin(w, r, req, "Please log in to purchase 2 or 4 credits.")
		return
	}

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
