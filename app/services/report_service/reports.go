// Package report_service stores purchased vehicle reports and fetches them from
// the configured report provider.
package report_service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/kernel"
	"carf-chorf.com/app/app/services/carbpi_service"
	"github.com/google/uuid"
)

// Dir returns the report storage directory.
func Dir() string {
	if path := strings.TrimSpace(kernel.ReportStorage.Path); path != "" {
		return path
	}
	return "var/reports"
}

// URL returns the absolute public report URL.
func URL(accessToken string) string {
	return strings.TrimRight(kernel.Server.Domain, "/") + "/reports/view?access_token=" + url.QueryEscape(accessToken)
}

// Save writes the report HTML for an access token.
func Save(accessToken string, html []byte) error {
	path, err := reportPath(accessToken)
	if err != nil {
		return err
	}
	if len(html) == 0 {
		return errors.New("report_service: empty report")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, html, 0o644)
}

// Read returns the report HTML for an access token.
func Read(accessToken string) ([]byte, error) {
	path, err := reportPath(accessToken)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

// Delete removes a stored report. A missing file is not an error.
func Delete(accessToken string) error {
	path, err := reportPath(accessToken)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// Has reports whether a stored report exists.
func Has(accessToken string) bool {
	path, err := reportPath(accessToken)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Prepare fetches the report HTML from the provider, preferring an existing
// archive and buying the report when no archive is available.
func Prepare(ctx context.Context, vin string) ([]byte, error) {
	html, err := carbpi_service.BuyReport(ctx, vin)
	if err != nil {
		return nil, err
	}
	if len(html) == 0 {
		return nil, errors.New("report_service: provider returned an empty report")
	}
	return html, nil
}

// Ensure returns stored HTML, fetching and saving it when missing.
func Ensure(ctx context.Context, order *domaine.Order) ([]byte, error) {
	if order == nil {
		return nil, errors.New("report_service: order is nil")
	}
	if Has(order.AccessToken) {
		return Read(order.AccessToken)
	}

	html, err := Prepare(ctx, order.Vin)
	if err != nil {
		return nil, err
	}
	if err := Save(order.AccessToken, html); err != nil {
		return nil, err
	}
	return html, nil
}

func reportPath(accessToken string) (string, error) {
	if len(accessToken) != 36 {
		return "", fmt.Errorf("report_service: invalid access token")
	}
	if _, err := uuid.Parse(accessToken); err != nil {
		return "", fmt.Errorf("report_service: invalid access token")
	}
	return filepath.Join(Dir(), accessToken+".html"), nil
}
