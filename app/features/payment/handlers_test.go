package payment

import (
	"testing"

	"carf-chorf.com/app/app/domaine"
)

func TestLabels(t *testing.T) {
	if got := statusLabel(domaine.TransactionPending); got != "Pending" {
		t.Fatalf("pending label = %q", got)
	}
	if got := statusLabel(domaine.TransactionPaid); got != "Paid" {
		t.Fatalf("paid label = %q", got)
	}
	if got := paymentMethodLabel(domaine.PaymentMethodPayPalCheckout); got != "PayPal Checkout" {
		t.Fatalf("payment method label = %q", got)
	}
}
