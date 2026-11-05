// Package reports renders purchased vehicle reports.
package reports

import (
	"bytes"
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
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(injectReportRuntime(html))
}

const reportRuntimeGuard = `<script>
(function () {
  var fetchImpl = window.fetch;
  window.fetch = function (input, init) {
    var url = typeof input === "string" ? input : (input && input.url) || "";
    if (url.indexOf("/fur/") !== -1 && /\/v1\.json(?:\?|$)/.test(url)) {
      return Promise.reject(new Error("report manifest disabled"));
    }
    return fetchImpl.apply(this, arguments);
  };
  try {
    var getItem = Storage.prototype.getItem;
    Storage.prototype.getItem = function (key) {
      if (typeof key === "string" && (key.indexOf("VHR") !== -1 || key.indexOf("fur") !== -1)) {
        return null;
      }
      return getItem.call(this, key);
    };
  } catch (e) {}
})();
</script><style>.report-header-print-button{display:none!important}</style><script src="/public/js/report-download.js"></script>`

func injectReportRuntime(html []byte) []byte {
	const marker = "</head>"
	if !bytes.Contains(html, []byte(marker)) {
		return html
	}
	return bytes.Replace(html, []byte(marker), append([]byte(reportRuntimeGuard), []byte(marker)...), 1)
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
