package payment

import (
	"errors"
	"testing"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/paypal_service"
)

func TestVerifyCapturedOrder(t *testing.T) {
	transaction := &domaine.Transaction{Amount: 3400}

	order := &paypal_service.Order{Status: "COMPLETED"}
	order.PurchaseUnits = append(order.PurchaseUnits, struct {
		Payments struct {
			Captures []paypal_service.Capture `json:"captures"`
		} `json:"payments"`
	}{})
	order.PurchaseUnits[0].Payments.Captures = []paypal_service.Capture{{
		Status: "COMPLETED",
	}}
	order.PurchaseUnits[0].Payments.Captures[0].Amount.CurrencyCode = "USD"
	order.PurchaseUnits[0].Payments.Captures[0].Amount.Value = "34.00"

	if err := verifyCapturedOrder(order, transaction); err != nil {
		t.Fatalf("verifyCapturedOrder() error = %v", err)
	}

	order.PurchaseUnits[0].Payments.Captures[0].Amount.Value = "35.00"
	if err := verifyCapturedOrder(order, transaction); !errors.Is(err, ErrPaymentAmountInvalid) {
		t.Fatalf("error = %v, want ErrPaymentAmountInvalid", err)
	}

	order.Status = "APPROVED"
	if err := verifyCapturedOrder(order, transaction); !errors.Is(err, ErrPaymentNotCompleted) {
		t.Fatalf("error = %v, want ErrPaymentNotCompleted", err)
	}
}
