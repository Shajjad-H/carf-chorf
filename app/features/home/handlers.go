package home

import (
	"net/http"

	"carf-chorf.com/app/app/features/vehicle"
	"carf-chorf.com/app/pkg/renderer"
)

// Index renders the home page and its vehicle lookup form.
func Index(w http.ResponseWriter, r *http.Request) {
	renderer.Pongo(w, r, "home/index.html", http.StatusOK, vehicle.FormContext())
}
