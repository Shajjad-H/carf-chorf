// Package formparse parses both URL-encoded and multipart HTTP form bodies.
package formparse

import (
	"errors"
	"net/http"
)

const maxMemory = 32 << 20

// Parse reads URL-encoded and multipart form bodies. ParseMultipartForm also
// parses URL-encoded forms, but returns ErrNotMultipart for them; that error is
// expected and ignored.
func Parse(r *http.Request) error {
	if err := r.ParseMultipartForm(maxMemory); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		return err
	}
	return nil
}
