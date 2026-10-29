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
}
