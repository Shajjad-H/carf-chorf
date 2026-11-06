package static

import (
	"net/http"
	"strings"

	"carf-chorf.com/app/app/kernel"
)

func Handler(root string) http.Handler {
	dir := http.Dir(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := dir.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		info, err := f.Stat()
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		if strings.HasSuffix(info.Name(), ".webmanifest") {
			w.Header().Set("Content-Type", "application/manifest+json")
		}

		if kernel.IsDev() {
			w.Header().Set("Cache-Control", "no-store")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}
		http.ServeContent(w, r, info.Name(), info.ModTime(), f)
	})
}
