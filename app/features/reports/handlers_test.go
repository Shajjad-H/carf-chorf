package reports

import (
	"bytes"
	"testing"
)

func TestInjectReportRuntime(t *testing.T) {
	html := []byte("<html><head></head><body></body></html>")
	out := injectReportRuntime(html)

	if !bytes.Contains(out, []byte("report manifest disabled")) {
		t.Fatal("runtime guard was not injected")
	}
	if !bytes.Contains(out, []byte("/public/js/report-download.js")) {
		t.Fatal("PDF helper script was not injected")
	}
	if !bytes.Contains(out, []byte("</head>")) {
		t.Fatal("head closing tag was lost")
	}
	if bytes.Count(out, []byte("</head>")) != 1 {
		t.Fatalf("head closing tag count = %d", bytes.Count(out, []byte("</head>")))
	}
}

func TestInjectReportRuntimeWithoutHead(t *testing.T) {
	html := []byte("<html><body>report</body></html>")
	out := injectReportRuntime(html)

	if !bytes.Equal(out, html) {
		t.Fatal("html without a head should be returned unchanged")
	}
}
