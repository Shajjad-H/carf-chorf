// Package paypal_service provides the PayPal Orders API client used by checkout.
package paypal_service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/http_client"
	"carf-chorf.com/app/pkg/money"
)

// Currency is the currency used for credit purchases.
const Currency = "USD"

// APIError describes a non-successful PayPal response.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("paypal: unexpected status %d", e.Status)
	}
	return fmt.Sprintf("paypal: status %d: %s", e.Status, e.Body)
}

// Capture is one capture in a PayPal order.
type Capture struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Amount struct {
		CurrencyCode string `json:"currency_code"`
		Value        string `json:"value"`
	} `json:"amount"`
	ProcessorResponse struct {
		ResponseCode string `json:"response_code"`
	} `json:"processor_response"`
}

// Order is the subset of a PayPal order response used by checkout.
type Order struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PurchaseUnits []struct {
		Payments struct {
			Captures []Capture `json:"captures"`
		} `json:"payments"`
	} `json:"purchase_units"`
}

// IsPaid reports whether the order completed with a completed capture.
func (o *Order) IsPaid() bool {
	if o == nil || o.Status != "COMPLETED" {
		return false
	}
	for _, unit := range o.PurchaseUnits {
		for _, capture := range unit.Payments.Captures {
			if capture.Status == "COMPLETED" {
				return true
			}
		}
	}
	return false
}

// CapturedAmount returns the currency and value of the first completed capture.
func (o *Order) CapturedAmount() (string, string, bool) {
	if o == nil {
		return "", "", false
	}
	for _, unit := range o.PurchaseUnits {
		for _, capture := range unit.Payments.Captures {
			if capture.Status == "COMPLETED" {
				return capture.Amount.CurrencyCode, capture.Amount.Value, true
			}
		}
	}
	return "", "", false
}

var processorMessages = map[string]string{
	"0500": "Your card was declined by the issuing bank.",
	"9500": "This card was flagged as potentially fraudulent.",
	"5400": "This card has expired.",
	"5180": "The card details appear to be invalid.",
	"5120": "The card has insufficient funds.",
	"9520": "This card was reported lost or stolen.",
	"1330": "The card account is not valid.",
	"5100": "The transaction was declined by the issuing bank.",
	"00N7": "The card security code is incorrect.",
}

// ErrorMessage returns a customer-facing message for a declined capture.
func (o *Order) ErrorMessage() string {
	if o == nil {
		return "The payment could not be completed."
	}
	for _, unit := range o.PurchaseUnits {
		for _, capture := range unit.Payments.Captures {
			if message, ok := processorMessages[capture.ProcessorResponse.ResponseCode]; ok {
				return message
			}
		}
	}
	return "The payment was declined. Please try another payment method."
}

// Client is a PayPal API client. The zero value is not usable; use NewClient.
type Client struct {
	baseURL   string
	clientID  string
	secret    string
	http      *http.Client
	tokenMu   sync.Mutex
	token     string
	expiresAt time.Time
}

// NewClient builds a PayPal client. A nil HTTP client uses the shared client.
func NewClient(baseURL, clientID, secret string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http_client.Shared
	}
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		clientID: clientID,
		secret:   secret,
		http:     httpClient,
	}
}

var (
	defaultClient *Client
	defaultOnce   sync.Once
)

func client() *Client {
	defaultOnce.Do(func() {
		defaultClient = NewClient(
			kernel.PayPal.APIURL,
			kernel.PayPal.ClientID,
			kernel.PayPal.ClientSecret,
			http_client.Shared,
		)
	})
	return defaultClient
}

// CreateOrder creates a PayPal order for the transaction using the default client.
func CreateOrder(ctx context.Context, transaction *domaine.Transaction) (string, error) {
	return client().CreateOrder(ctx, transaction)
}

// CreateOrder creates a PayPal order for the transaction.
func (c *Client) CreateOrder(ctx context.Context, transaction *domaine.Transaction) (string, error) {
	payload := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{
			{
				"invoice_id":   transaction.Token,
				"reference_id": transaction.Token,
				"description":  fmt.Sprintf("%d CARFAX report credit(s)", transaction.Credits),
				"amount": map[string]string{
					"currency_code": Currency,
					"value":         money.Decimal(transaction.Amount),
				},
			},
		},
		"application_context": map[string]string{
			"brand_name":          "Carf Chorf",
			"locale":              "en-US",
			"shipping_preference": "NO_SHIPPING",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	var response struct {
		ID string `json:"id"`
	}
	if err := c.do(ctx, http.MethodPost, "/v2/checkout/orders", body, &response); err != nil {
		return "", err
	}
	if response.ID == "" {
		return "", fmt.Errorf("paypal: create order returned an empty id")
	}
	return response.ID, nil
}

// CaptureOrder captures an approved PayPal order using the default client.
func CaptureOrder(ctx context.Context, pid string) (*Order, error) {
	return client().CaptureOrder(ctx, pid)
}

// CaptureOrder captures an approved PayPal order.
func (c *Client) CaptureOrder(ctx context.Context, pid string) (*Order, error) {
	var order Order
	path := "/v2/checkout/orders/" + url.PathEscape(pid) + "/capture"
	if err := c.do(ctx, http.MethodPost, path, nil, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

// GetOrder fetches a PayPal order using the default client.
func GetOrder(ctx context.Context, pid string) (*Order, error) {
	return client().GetOrder(ctx, pid)
}

// GetOrder fetches a PayPal order.
func (c *Client) GetOrder(ctx context.Context, pid string) (*Order, error) {
	var order Order
	path := "/v2/checkout/orders/" + url.PathEscape(pid)
	if err := c.do(ctx, http.MethodGet, path, nil, &order); err != nil {
		return nil, err
	}
	return &order, nil
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}

	body := bytes.NewBufferString("grant_type=client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/oauth2/token", body)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.clientID, c.secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &APIError{Status: resp.StatusCode, Body: strings.TrimSpace(string(raw))}
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(raw, &tokenResponse); err != nil {
		return "", fmt.Errorf("paypal: decode token: %w", err)
	}
	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("paypal: empty access token")
	}

	ttl := time.Duration(tokenResponse.ExpiresIn) * time.Second
	if ttl > time.Minute {
		ttl -= time.Minute
	}
	c.token = tokenResponse.AccessToken
	c.expiresAt = time.Now().Add(ttl)
	return c.token, nil
}

func (c *Client) do(ctx context.Context, method, path string, body []byte, target any) error {
	token, err := c.accessToken(ctx)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{Status: resp.StatusCode, Body: strings.TrimSpace(string(raw))}
	}
	if target == nil {
		return nil
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("paypal: decode response: %w", err)
	}
	return nil
}
