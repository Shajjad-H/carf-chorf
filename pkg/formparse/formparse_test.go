package formparse

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestParseMultipart(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("vin", "3GNCJKSBXLL236612"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := Parse(req); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := req.PostFormValue("vin"); got != "3GNCJKSBXLL236612" {
		t.Fatalf("vin = %q", got)
	}
}

func TestParseURLEncoded(t *testing.T) {
	values := url.Values{"vin": {"3GNCJKSBXLL236612"}}
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if err := Parse(req); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := req.PostFormValue("vin"); got != "3GNCJKSBXLL236612" {
		t.Fatalf("vin = %q", got)
	}
}
