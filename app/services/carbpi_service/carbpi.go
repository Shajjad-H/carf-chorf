// Package carbpi_service provides the client for the carbpi.fr vehicle-history API.
package carbpi_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/pkg/http_client"
	"carf-chorf.com/app/pkg/simple_logger"
)

// BaseURL is the production base URL for the carbpi API.
const BaseURL = "https://carbpi.fr/api/v1"

// Record is the subset of a carbpi vehicle record used by the lookup flow.
type Record struct {
	Vin           string
	Plate         string
	State         string
	Vehicle       string
	CarfaxRecords uint
}

// APIError describes a non-200 carbpi response.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("carbpi: %s (status %d)", e.Message, e.Status)
	}
	return fmt.Sprintf("carbpi: unexpected status %d", e.Status)
}

// ErrNoVIN indicates that carbpi returned a record without a usable VIN.
var ErrNoVIN = errors.New("carbpi: no vin in record response")

// IsNotFound reports whether err is a carbpi 404 response.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.Status == http.StatusNotFound
}

// Client is an authenticated carbpi API client. The zero value is not usable;
// construct one with NewClient.
type Client struct {
	baseURL     string
	tokenName   string
	accessToken string
	http        *http.Client
}

// NewClient builds a carbpi client. A nil HTTP client uses the shared client.
func NewClient(baseURL, tokenName, accessToken string, client *http.Client) *Client {
	if client == nil {
		client = http_client.Shared
	}

	return &Client{
		baseURL:     strings.TrimRight(baseURL, "/"),
		tokenName:   tokenName,
		accessToken: accessToken,
		http:        client,
	}
}

// The default client is lazy so tests can exercise the package without
// initializing application config.
var (
	defaultClient *Client
	defaultOnce   sync.Once
)

func client() *Client {
	defaultOnce.Do(func() {
		defaultClient = NewClient(
			BaseURL,
			kernel.ReportProvider.CarbpiTokenName,
			kernel.ReportProvider.CarbpiTokenSecret,
			http_client.Shared,
		)
	})
	return defaultClient
}

// CheckCarfaxRecords fetches a vehicle record by VIN using the default client.
func CheckCarfaxRecords(ctx context.Context, vin string) (*Record, error) {
	return client().CheckCarfaxRecords(ctx, vin)
}

// CheckCarfaxRecords fetches a vehicle record by VIN.
func (c *Client) CheckCarfaxRecords(ctx context.Context, vin string) (*Record, error) {
	simple_logger.Debugf(kernel.IsDev(), "carbpi: checking carfax records for vin: %s", vin)
	return c.getRecord(ctx, "/record/carfax/"+url.PathEscape(vin))
}

// CheckCarfaxRecordsByPlate fetches a vehicle record by plate and state using
// the default client.
func CheckCarfaxRecordsByPlate(ctx context.Context, plate, state string) (*Record, error) {
	return client().CheckCarfaxRecordsByPlate(ctx, plate, state)
}

// CheckCarfaxRecordsByPlate fetches a vehicle record by plate and state. The
// response includes the VIN, so no separate VIN-resolution call is needed.
func (c *Client) CheckCarfaxRecordsByPlate(ctx context.Context, plate, state string) (*Record, error) {
	simple_logger.Debugf(kernel.IsDev(), "carbpi: checking carfax records for plate: %s state: %s", plate, state)
	path := "/record/carfax/" + url.PathEscape(plate) + "/" + url.PathEscape(state)
	return c.getRecord(ctx, path)
}

// getRecord calls one of the two record endpoints; both return the same shape.
// carfax_reports is nullable upstream, but decoding null into uint leaves it at
// zero, which is the value the UI uses for "no records".
func (c *Client) getRecord(ctx context.Context, path string) (*Record, error) {
	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var response struct {
		Vin           string `json:"vin"`
		Plate         string `json:"plate"`
		State         string `json:"state"`
		Model         string `json:"model"`
		CarfaxReports uint   `json:"carfax_reports"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("carbpi: decode record: %w", err)
	}

	// Both lookup paths need a usable VIN; the package page and cache key
	// depend on it, so reject malformed provider responses.
	vin := strings.ToUpper(strings.TrimSpace(response.Vin))
	if len(vin) != 17 {
		return nil, ErrNoVIN
	}

	return &Record{
		Vin:           vin,
		Plate:         strings.ToUpper(strings.TrimSpace(response.Plate)),
		State:         strings.ToUpper(strings.TrimSpace(response.State)),
		Vehicle:       strings.TrimSpace(response.Model),
		CarfaxRecords: response.CarfaxReports,
	}, nil
}

// GetArchiveReport fetches an already purchased CARFAX report using the
// default client.
func GetArchiveReport(ctx context.Context, vin string) ([]byte, error) {
	return client().GetArchiveReport(ctx, vin)
}

// GetArchiveReport fetches an already purchased CARFAX report.
func (c *Client) GetArchiveReport(ctx context.Context, vin string) ([]byte, error) {
	simple_logger.Debugf(kernel.IsDev(), "carbpi: fetching archived report for vin: %s", vin)
	return c.get(ctx, "/report/carfax/"+url.PathEscape(vin)+"/archive")
}

// BuyReport purchases a CARFAX report using the default client.
func BuyReport(ctx context.Context, vin string) ([]byte, error) {
	return client().BuyReport(ctx, vin)
}

// BuyReport purchases a CARFAX report and returns its HTML.
func (c *Client) BuyReport(ctx context.Context, vin string) ([]byte, error) {
	simple_logger.Debugf(kernel.IsDev(), "carbpi: buying report for vin: %s", vin)
	return c.get(ctx, "/report/carfax/"+url.PathEscape(vin))
}

// get performs an authenticated GET and converts non-200 responses into APIError.
func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Token-Name", c.tokenName)
	req.Header.Set("X-Access-Token", c.accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errorFromBody(body, resp.StatusCode)
	}

	return body, nil
}

// errorFromBody prefers carbpi's JSON {"error": "..."} message, falling back
// to the raw response body for unexpected error pages.
func errorFromBody(body []byte, status int) error {
	var errResp struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return &APIError{Status: status, Message: errResp.Error}
	}

	return &APIError{Status: status, Message: strings.TrimSpace(string(body))}
}
