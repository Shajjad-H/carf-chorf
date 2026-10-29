package paypal_service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"carf-chorf.com/app/app/domaine"
)

func TestCreateAndCaptureOrder(t *testing.T) {
	var tokenCalls int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/oauth2/token":
			atomic.AddInt32(&tokenCalls, 1)
			user, pass, ok := r.BasicAuth()
			if !ok || user != "client" || pass != "secret" {
				t.Fatalf("unexpected basic auth: %q/%q", user, pass)
			}
			_, _ = w.Write([]byte(`{"access_token":"token","expires_in":3600}`))
		case "/v2/checkout/orders":
			if got := r.Header.Get("Authorization"); got != "Bearer token" {
				t.Fatalf("Authorization = %q", got)
			}
			var payload struct {
				Intent        string `json:"intent"`
				PurchaseUnits []struct {
					InvoiceID string `json:"invoice_id"`
					Amount    struct {
						CurrencyCode string `json:"currency_code"`
						Value        string `json:"value"`
					} `json:"amount"`
				} `json:"purchase_units"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode create order: %v", err)
			}
			if payload.Intent != "CAPTURE" {
				t.Fatalf("intent = %q", payload.Intent)
			}
			if len(payload.PurchaseUnits) != 1 {
				t.Fatalf("purchase units = %d", len(payload.PurchaseUnits))
			}
			if payload.PurchaseUnits[0].InvoiceID != "transaction-token" {
				t.Fatalf("invoice id = %q", payload.PurchaseUnits[0].InvoiceID)
			}
			if payload.PurchaseUnits[0].Amount.CurrencyCode != "USD" || payload.PurchaseUnits[0].Amount.Value != "34.00" {
				t.Fatalf("amount = %+v", payload.PurchaseUnits[0].Amount)
			}
			_, _ = w.Write([]byte(`{"id":"order-123"}`))
		case "/v2/checkout/orders/order-123/capture":
			_, _ = w.Write([]byte(`{
				"id":"order-123",
				"status":"COMPLETED",
				"purchase_units":[{"payments":{"captures":[{
					"id":"capture-1",
					"status":"COMPLETED",
					"amount":{"currency_code":"USD","value":"34.00"}
				}]}}]
			}`))
		case "/v2/checkout/orders/order-123":
			_, _ = w.Write([]byte(`{"id":"order-123","status":"APPROVED"}`))
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "client", "secret", server.Client())
	transaction := &domaine.Transaction{Token: "transaction-token", Credits: 2, Amount: 3400}

	pid, err := client.CreateOrder(context.Background(), transaction)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if pid != "order-123" {
		t.Fatalf("pid = %q", pid)
	}

	order, err := client.CaptureOrder(context.Background(), pid)
	if err != nil {
		t.Fatalf("CaptureOrder() error = %v", err)
	}
	if !order.IsPaid() {
		t.Fatal("expected paid order")
	}
	currency, value, ok := order.CapturedAmount()
	if !ok || currency != "USD" || value != "34.00" {
		t.Fatalf("captured amount = %q/%q/%t", currency, value, ok)
	}

	if _, err := client.GetOrder(context.Background(), pid); err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}

	if got := atomic.LoadInt32(&tokenCalls); got != 1 {
		t.Fatalf("token calls = %d, want 1", got)
	}
}

func TestOrderDeclineMessage(t *testing.T) {
	order := &Order{Status: "COMPLETED"}
	order.PurchaseUnits = append(order.PurchaseUnits, struct {
		Payments struct {
			Captures []Capture `json:"captures"`
		} `json:"payments"`
	}{})
	order.PurchaseUnits[0].Payments.Captures = []Capture{{
		Status: "DECLINED",
	}}
	order.PurchaseUnits[0].Payments.Captures[0].ProcessorResponse.ResponseCode = "5400"

	if got := order.ErrorMessage(); !strings.Contains(got, "expired") {
		t.Fatalf("message = %q", got)
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{Status: 500, Body: "boom"}
	if got := err.Error(); !strings.Contains(got, "500") || !strings.Contains(got, "boom") {
		t.Fatalf("error = %q", got)
	}
}
