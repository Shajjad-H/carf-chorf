// Package reports renders purchased vehicle reports.
package reports

import (
	"net/http"
	"strings"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/report_service"
	"carf-chorf.com/app/app/services/transaction_service"
	"carf-chorf.com/app/pkg/renderer"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/flosch/pongo2/v7"
)

// View renders the page that hosts a purchased report.
func View(w http.ResponseWriter, r *http.Request) {
	order, ok := findOrder(w, r)
	if !ok {
		return
	}

	if _, err := report_service.Ensure(r.Context(), order); err != nil {
		simple_logger.Errorf("reports: load report: %v", err)
		http.Error(w, "The report could not be loaded. Please try again later.", http.StatusBadGateway)
		return
	}

	renderer.Pongo(w, r, "reports/show.html", http.StatusOK, pongo2.Context{
		"order": order,
	})
}

// Content serves the raw report HTML.
func Content(w http.ResponseWriter, r *http.Request) {
	order, ok := findOrder(w, r)
	if !ok {
		return
	}

	html, err := report_service.Ensure(r.Context(), order)
	if err != nil {
		simple_logger.Errorf("reports: load report content: %v", err)
		http.Error(w, "The report could not be loaded. Please try again later.", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(html)
}

func findOrder(w http.ResponseWriter, r *http.Request) (*domaine.Order, bool) {
	token := strings.TrimSpace(r.URL.Query().Get("access_token"))
	order, err := transaction_service.FindOrderByAccessToken(token)
	if err != nil {
		simple_logger.Errorf("reports: find order: %v", err)
		http.Error(w, "The report could not be loaded. Please try again later.", http.StatusInternalServerError)
		return nil, false
	}
	if order == nil {
		http.Error(w, "Report not found.", http.StatusNotFound)
		return nil, false
	}
	return order, true
}
