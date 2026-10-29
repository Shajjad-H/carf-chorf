// Package dashboard implements the authenticated account dashboard.
package dashboard

import (
	"errors"
	"net/http"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/features/vehicle"
	auth_service "carf-chorf.com/app/app/services/auth"
	"carf-chorf.com/app/app/services/carbpi_service"
	"carf-chorf.com/app/app/services/order_service"
	"carf-chorf.com/app/app/services/transaction_service"
	"carf-chorf.com/app/app/services/user_service"
	dto "carf-chorf.com/app/pkg/dto_validations"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/response"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
)

var (
	ErrUnauthenticated = errors.New("not authenticated")
	ErrInactiveUser    = errors.New("account is not active")
)

const (
	indexTemplate        = "dashboard/index.html"
	ordersTemplate       = "dashboard/orders.html"
	transactionsTemplate = "dashboard/transactions.html"
)

// Index renders the dashboard with credits, the report purchase form, and the
// user's report history.
func Index(w http.ResponseWriter, r *http.Request) {
	user, err := loadUser(r)
	if err != nil {
		handleLoadUserError(w, r, err)
		return
	}

	orders, err := order_service.ListByUser(user.ID)
	if err != nil {
		simple_logger.Errorf("dashboard: list orders: %v", err)
		http.Error(w, "Something went wrong. Please try again.", http.StatusInternalServerError)
		return
	}

	renderer.Pongo(w, r, indexTemplate, http.StatusOK, pongo2.Context{
		"credits": user.Credits,
		"email":   user.Email,
		"states":  vehicle.States(),
		"orders":  orders,
	})
}

// Orders renders the user's purchased reports.
func Orders(w http.ResponseWriter, r *http.Request) {
	user, err := loadUser(r)
	if err != nil {
		handleLoadUserError(w, r, err)
		return
	}

	orders, err := order_service.ListByUser(user.ID)
	if err != nil {
		simple_logger.Errorf("dashboard: list orders: %v", err)
		http.Error(w, "Something went wrong. Please try again.", http.StatusInternalServerError)
		return
	}

	renderer.Pongo(w, r, ordersTemplate, http.StatusOK, pongo2.Context{
		"orders":      orders,
		"order_count": len(orders),
	})
}

// Transactions renders the user's completed credit purchases.
func Transactions(w http.ResponseWriter, r *http.Request) {
	user, err := loadUser(r)
	if err != nil {
		handleLoadUserError(w, r, err)
		return
	}

	transactions, err := transaction_service.ListPaidByUser(user.ID)
	if err != nil {
		simple_logger.Errorf("dashboard: list paid transactions: %v", err)
		http.Error(w, "Something went wrong. Please try again.", http.StatusInternalServerError)
		return
	}

	var totalSpent, totalCredits uint
	for _, transaction := range transactions {
		totalSpent += transaction.Amount
		totalCredits += transaction.Credits
	}

	renderer.Pongo(w, r, transactionsTemplate, http.StatusOK, pongo2.Context{
		"transactions":      transactions,
		"transaction_count": len(transactions),
		"total_spent":       totalSpent,
		"total_credits":     totalCredits,
	})
}

// LegacyHistory redirects the old combined history page to reports.
func LegacyHistory(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/dashboard/orders", http.StatusSeeOther)
}

// Lookup resolves a vehicle record for the confirmation modal without spending
// a credit.
func Lookup(w http.ResponseWriter, r *http.Request) {
	user, err := loadUser(r)
	if err != nil {
		handleJSONUserError(w, err)
		return
	}
	if user.Credits < 1 {
		response.Error(w, http.StatusPaymentRequired, "You don't have enough credits. Please buy more credits.")
		return
	}

	if err := r.ParseForm(); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid form submission.")
		return
	}

	req := dto.LookupRequest{
		Vin:   r.PostFormValue("vin"),
		Plate: r.PostFormValue("plate"),
		State: r.PostFormValue("state"),
	}
	if err := req.Validate(); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "Please enter a valid VIN or a plate and state.")
		return
	}

	record, err := vehicle.Resolve(r.Context(), req)
	if err != nil {
		writeLookupError(w, err)
		return
	}
	if record.CarfaxRecords == 0 {
		response.Error(w, http.StatusNotFound, "No records available for this vehicle.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"token":   record.Token,
		"vehicle": record.Vehicle,
		"vin":     record.Vin,
		"plate":   record.Plate,
		"state":   record.State,
		"records": record.CarfaxRecords,
	})
}

// Purchase consumes one credit and creates a report order.
func Purchase(w http.ResponseWriter, r *http.Request) {
	user, err := loadUser(r)
	if err != nil {
		handleJSONUserError(w, err)
		return
	}

	if err := r.ParseForm(); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid form submission.")
		return
	}

	record, err := vehicle.FindRecordByToken(r.PostFormValue("token"))
	if err != nil {
		simple_logger.Errorf("dashboard: find record: %v", err)
		response.Error(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
		return
	}
	if record == nil {
		response.Error(w, http.StatusNotFound, "Vehicle record not found. Please run a new lookup.")
		return
	}

	order, err := order_service.BuyReportWithCredit(r.Context(), user.ID, record)
	if err != nil {
		if errors.Is(err, order_service.ErrNotEnoughCredits) {
			response.Error(w, http.StatusPaymentRequired, "You don't have enough credits. Please buy more credits.")
			return
		}
		simple_logger.Errorf("dashboard: buy report: %v", err)
		response.Error(w, http.StatusBadGateway, "We could not prepare the report. Please try again.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"redirect": "/reports/view?access_token=" + order.AccessToken,
	})
}

func writeLookupError(w http.ResponseWriter, err error) {
	switch {
	case carbpi_service.IsNotFound(err):
		response.Error(w, http.StatusNotFound, "We couldn't find a vehicle for that VIN or plate and state.")
	case errors.Is(err, carbpi_service.ErrNoVIN):
		response.Error(w, http.StatusUnprocessableEntity, "We couldn't verify that vehicle. Please check the VIN or plate and state.")
	default:
		simple_logger.Errorf("dashboard: resolve vehicle: %v", err)
		response.Error(w, http.StatusBadGateway, "Vehicle lookup is temporarily unavailable. Please try again.")
	}
}

func loadUser(r *http.Request) (*domaine.User, error) {
	state, ok := auth_service.Current(r)
	if !ok {
		return nil, ErrUnauthenticated
	}

	user, err := user_service.FindByID(state.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrInactiveUser
	}
	return user, nil
}

func handleLoadUserError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, ErrInactiveUser) {
		_ = auth_service.ClearCurrent(w, r)
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func handleJSONUserError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrUnauthenticated) {
		response.Error(w, http.StatusUnauthorized, "Please log in.")
		return
	}
	response.Error(w, http.StatusUnauthorized, "Your account is not active.")
}
