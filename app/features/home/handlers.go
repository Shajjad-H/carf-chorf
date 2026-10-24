package home

import (
	"net/http"

	"carf-chorf.com/app/pkg/renderer"
)

func Index(w http.ResponseWriter, r *http.Request) {
	renderer.Pongo(w, r, "home/index.html", http.StatusOK, nil)
}
