// Package payment renders the pending transaction summary and payment stub.
package payment

import (
	"net/http"
	"strings"

	"carf-chorf.com/app/app/domaine"
	flashes_service "carf-chorf.com/app/app/services/flashes"
	"carf-chorf.com/app/app/services/transaction_service"
	"carf-chorf.com/app/pkg/money"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
)

const templateName = "payment/show.html"

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
		"credits":        transaction.Credits,
		"amount":         money.FormatCents(transaction.Amount),
		"status":         statusLabel(transaction.Status),
		"payment_method": paymentMethodLabel(transaction.PaymentMethod),
		"email":          transaction.User.Email,
		"has_vehicle":    transaction.CacheRecord != nil,
		"is_pending":     transaction.Status == domaine.TransactionPending,
	}
	if transaction.CacheRecord != nil {
		ctx["vehicle"] = transaction.CacheRecord.Vehicle
		ctx["vin"] = transaction.CacheRecord.Vin
		ctx["plate"] = transaction.CacheRecord.Plate
		ctx["state"] = transaction.CacheRecord.State
	}

	renderer.Pongo(w, r, templateName, http.StatusOK, ctx)
}

// statusLabel returns the display label for a transaction status.
func statusLabel(status domaine.TransactionStatus) string {
	if status == domaine.TransactionPaid {
		return "Paid"
	}
	return "Pending"
}

// paymentMethodLabel returns the display label for a payment method.
func paymentMethodLabel(method domaine.PaymentMethod) string {
	if method == domaine.PaymentMethodPayPalCheckout {
		return "PayPal Checkout"
	}
	return string(method)
}
