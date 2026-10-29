// Package payment renders payment pages and handles PayPal checkout endpoints.
package payment

import (
	"errors"
	"net/http"
	"strings"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/kernel"
	flashes_service "carf-chorf.com/app/app/services/flashes"
	"carf-chorf.com/app/app/services/transaction_service"
	"carf-chorf.com/app/pkg/money"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/response"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
)

const (
	templateName        = "payment/show.html"
	successTemplateName = "payment/success.html"
)

// Show renders the payment page for a transaction token.
func Show(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))

	transaction, err := transaction_service.FindByToken(token)
	if err != nil {
		simple_logger.Errorf("payment: find transaction: %v", err)
		flashes_service.AddFlash(w, r, "error", "Transaction not found. Please choose a package.")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}
	if transaction == nil {
		flashes_service.AddFlash(w, r, "error", "Transaction not found. Please choose a package.")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	ctx := pongo2.Context{
		"token":            transaction.Token,
		"credits":          transaction.Credits,
		"amount":           money.FormatCents(transaction.Amount),
		"status":           statusLabel(transaction.Status),
		"email":            transaction.User.Email,
		"has_vehicle":      transaction.CacheRecord != nil,
		"is_pending":       transaction.Status == domaine.TransactionPending,
		"paypal_client_id": kernel.PayPal.ClientID,
	}
	if transaction.CacheRecord != nil {
		ctx["vehicle"] = transaction.CacheRecord.Vehicle
		ctx["vin"] = transaction.CacheRecord.Vin
		ctx["plate"] = transaction.CacheRecord.Plate
		ctx["state"] = transaction.CacheRecord.State
	}

	renderer.Pongo(w, r, templateName, http.StatusOK, ctx)
}

// Success renders the post-payment page for transactions without a vehicle.
func Success(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))

	transaction, err := transaction_service.FindByToken(token)
	if err != nil {
		simple_logger.Errorf("payment: find transaction: %v", err)
		flashes_service.AddFlash(w, r, "error", "Transaction not found. Please choose a package.")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}
	if transaction == nil || transaction.Status != domaine.TransactionPaid {
		flashes_service.AddFlash(w, r, "error", "Payment has not been completed.")
		http.Redirect(w, r, "/pricing", http.StatusSeeOther)
		return
	}

	renderer.Pongo(w, r, successTemplateName, http.StatusOK, pongo2.Context{
		"credits": transaction.Credits,
		"amount":  money.FormatCents(transaction.Amount),
		"email":   transaction.User.Email,
	})
}

// CreateOrder creates a PayPal order for a pending transaction.
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	pid, err := CreatePayPalOrder(r.Context(), strings.TrimSpace(r.URL.Query().Get("token")))
	if err != nil {
		writePaymentError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"pid": pid})
}

// CaptureOrder captures a PayPal order and completes the purchase.
func CaptureOrder(w http.ResponseWriter, r *http.Request) {
	redirect, err := CapturePayPalOrder(r.Context(), strings.TrimSpace(r.URL.Query().Get("token")))
	if err != nil {
		writePaymentError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"redirect": redirect,
	})
}

func writePaymentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrTransactionNotFound):
		response.Error(w, http.StatusNotFound, "Transaction not found.")
	case errors.Is(err, ErrPaymentNotStarted):
		response.Error(w, http.StatusBadRequest, "Payment has not been started.")
	case errors.Is(err, ErrPaymentNotCompleted), errors.Is(err, ErrPaymentAmountInvalid):
		response.Error(w, http.StatusBadRequest, "The payment was not completed.")
	case errors.Is(err, transaction_service.ErrTransactionNotPending):
		response.Error(w, http.StatusBadRequest, "This transaction is already paid.")
	default:
		simple_logger.Errorf("payment: %v", err)
		response.Error(w, http.StatusBadGateway, "Payment could not be completed. Please try again.")
	}
}

// statusLabel returns the display label for a transaction status.
func statusLabel(status domaine.TransactionStatus) string {
	if status == domaine.TransactionPaid {
		return "Paid"
	}
	return "Pending"
}
