package home

import (
	"net/http"

	"carf-chorf.com/app/app/features/checkout"
	"carf-chorf.com/app/app/features/vehicle"
	auth_service "carf-chorf.com/app/app/services/auth"
	"carf-chorf.com/app/pkg/renderer"
)

// Index renders the home page and its vehicle lookup form. Signed-in users are
// sent to their dashboard instead.
func Index(w http.ResponseWriter, r *http.Request) {
	if auth_service.IsAuthenticated(r) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	ctx := vehicle.FormContext()
	ctx["packages"] = checkout.PackageViews()
	renderer.Pongo(w, r, "home/index.html", http.StatusOK, ctx)
}
