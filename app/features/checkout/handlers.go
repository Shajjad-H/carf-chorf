// Package checkout implements credit package selection and transaction creation.
package checkout

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/features/vehicle"
	auth_service "carf-chorf.com/app/app/services/auth"
	flashes_service "carf-chorf.com/app/app/services/flashes"
	"carf-chorf.com/app/app/services/user_service"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
)

const templateName = "checkout/index.html"

// Index renders the checkout page for a vehicle lookup or for a logged-in user
// buying credits directly.
func Index(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	requestedCredits := parseCredits(r.URL.Query().Get("credits"))

	record, err := loadCheckoutVehicle(token)
	if err != nil {
		if !errors.Is(err, ErrVehicleNotFound) {
			simple_logger.Errorf("checkout: find vehicle: %v", err)
		}
		redirectWithFlash(w, r, "/", "Vehicle record not found. Please run a new lookup.")
		return
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
	renderer.Pongo(w, r, templateName, http.StatusOK, ctx)
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

func loadCheckoutVehicle(token string) (*domaine.CacheRecord, error) {
	if token == "" {
		return nil, nil
	}

	record, err := vehicle.FindRecordByToken(token)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrVehicleNotFound
	}
	return record, nil
}

func loadCheckoutUser(r *http.Request) (*domaine.User, bool, error) {
	state, authenticated := auth_service.Current(r)
	if !authenticated {
		return nil, false, nil
	}

	user, err := user_service.FindByID(state.UserID)
	if err != nil {
		return nil, true, err
	}
	if user == nil || !user.IsActive {
		return nil, true, ErrInactiveUser
	}
	return user, true, nil
}

func selectedCheckoutCredits(requested uint, record *domaine.CacheRecord, authenticated bool) uint {
	if authenticated {
		return requested
	}
	if requested == 1 && record != nil {
		return 1
	}
	return 0
}

func checkoutPageContext(token string, requested, selected uint, record *domaine.CacheRecord, user *domaine.User) pongo2.Context {
	ctx := pongo2.Context{
		"packages":          PackageViews(),
		"selected_credits":  selected,
		"requested_credits": requested,
		"token":             token,
		"has_vehicle":       record != nil,
		"user_credits":      uint(0),
	}

	if record != nil {
		ctx["vehicle"] = record.Vehicle
		ctx["vin"] = record.Vin
		ctx["plate"] = record.Plate
		ctx["state"] = record.State
		ctx["records"] = record.CarfaxRecords
	}
	if user != nil {
		ctx["user_credits"] = user.Credits
	}

	return ctx
}

func parsePurchaseRequest(r *http.Request) PurchaseRequest {
	req := PurchaseRequest{
		Credits:       parseCredits(r.PostFormValue("credits")),
		Email:         r.PostFormValue("email"),
		ConfirmEmail:  r.PostFormValue("confirm_email"),
		Token:         r.PostFormValue("token"),
		TermsAccepted: r.PostFormValue("terms") == "on",
	}
	req.normalize()
	return req
}

func handlePurchaseError(w http.ResponseWriter, r *http.Request, req PurchaseRequest, authenticated bool, err error) {
	switch {
	case errors.Is(err, ErrInactiveUser) && authenticated:
		_ = auth_service.ClearCurrent(w, r)
		redirectToLogin(w, r, req, "Please log in again.")
	case errors.Is(err, ErrGuestCredits):
		redirectToLogin(w, r, req, "Please log in to purchase 2 or 4 credits.")
	case errors.Is(err, ErrVehicleNotFound):
		redirectWithFlash(w, r, "/", "Vehicle record not found. Please run a new lookup.")
	default:
		simple_logger.Errorf("checkout: create purchase: %v", err)
		redirectWithFlash(w, r, buildCheckoutURL(req.Token, req.Credits), purchaseErrorMessage(err))
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request, req PurchaseRequest, message string) {
	next := buildCheckoutURL(req.Token, req.Credits)
	redirectWithFlash(w, r, "/login?next="+url.QueryEscape(next), message)
}

func redirectWithFlash(w http.ResponseWriter, r *http.Request, target, message string) {
	flashes_service.AddFlash(w, r, "error", message)
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// parseCredits returns 0 unless the value matches a catalog package.
func parseCredits(raw string) uint {
	credits, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 32)
	if err != nil {
		return 0
	}
	if _, ok := Find(uint(credits)); !ok {
		return 0
	}
	return uint(credits)
}

// buildCheckoutURL preserves the vehicle token and selected credits across
// validation errors and login redirects.
func buildCheckoutURL(token string, credits uint) string {
	query := url.Values{}
	if token != "" {
		query.Set("token", token)
	}
	if credits != 0 {
		query.Set("credits", strconv.FormatUint(uint64(credits), 10))
	}
	if len(query) == 0 {
		return "/checkout"
	}
	return "/checkout?" + query.Encode()
}

// purchaseErrorMessage maps internal checkout errors to user-facing copy.
func purchaseErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrVehicleRequired):
		return "Start with a vehicle lookup to buy a single report credit."
	case errors.Is(err, ErrVehicleNotFound):
		return "Vehicle record not found. Please run a new lookup."
	case errors.Is(err, ErrInvalidEmail):
		return "Enter a valid email address."
	case errors.Is(err, ErrEmailMismatch):
		return "Email addresses do not match."
	case errors.Is(err, ErrTermsRequired):
		return "Please accept the Terms and Conditions."
	case errors.Is(err, ErrInactiveUser):
		return "This account is not active."
	case errors.Is(err, ErrInvalidPackage):
		return "Please select a valid package."
	default:
		return "Something went wrong. Please try again."
	}
}
