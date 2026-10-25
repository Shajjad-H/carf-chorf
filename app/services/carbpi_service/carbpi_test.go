package carbpi_service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckCarfaxRecords(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/record/carfax/1HGCM82633A004352" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("X-Token-Name"); got != "name" {
			t.Fatalf("X-Token-Name = %q", got)
		}
		if got := r.Header.Get("X-Access-Token"); got != "secret" {
			t.Fatalf("X-Access-Token = %q", got)
		}
		_, _ = io.WriteString(w, `{
			"vin":"1HGCM82633A004352",
			"plate":"ABC123",
			"state":"CA",
			"model":"2003 Honda Accord",
			"carfax_reports":4,
			"record_provider":"simulcast",
			"vin_provider":"reports_vin"
		}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "name", "secret", server.Client())
	record, err := client.CheckCarfaxRecords(context.Background(), "1HGCM82633A004352")
	if err != nil {
		t.Fatalf("CheckCarfaxRecords() error = %v", err)
	}
	if record.Vin != "1HGCM82633A004352" {
		t.Fatalf("Vin = %q", record.Vin)
	}
	if record.Plate != "ABC123" || record.State != "CA" {
		t.Fatalf("plate/state = %q/%q", record.Plate, record.State)
	}
	if record.Vehicle != "2003 Honda Accord" {
		t.Fatalf("Vehicle = %q", record.Vehicle)
	}
	if record.CarfaxRecords != 4 {
		t.Fatalf("CarfaxRecords = %d", record.CarfaxRecords)
	}
}

func TestCheckCarfaxRecordsByPlate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/record/carfax/ABC123/CA" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{
			"vin":"1HGCM82633A004352",
			"plate":"ABC123",
			"state":"CA",
			"model":"2003 Honda Accord",
			"carfax_reports":3
		}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "name", "secret", server.Client())
	record, err := client.CheckCarfaxRecordsByPlate(context.Background(), "ABC123", "CA")
	if err != nil {
		t.Fatalf("CheckCarfaxRecordsByPlate() error = %v", err)
	}
	if record.Vin != "1HGCM82633A004352" {
		t.Fatalf("Vin = %q", record.Vin)
	}
	if record.CarfaxRecords != 3 {
		t.Fatalf("CarfaxRecords = %d", record.CarfaxRecords)
	}
}

func TestCheckCarfaxRecordsNullReports(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"vin":"1HGCM82633A004352","model":"2003 Honda Accord","carfax_reports":null}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "name", "secret", server.Client())
	record, err := client.CheckCarfaxRecords(context.Background(), "1HGCM82633A004352")
	if err != nil {
		t.Fatalf("CheckCarfaxRecords() error = %v", err)
	}
	if record.CarfaxRecords != 0 {
		t.Fatalf("CarfaxRecords = %d, want 0", record.CarfaxRecords)
	}
}

func TestCheckCarfaxRecordsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"error":"record not found"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "name", "secret", server.Client())
	_, err := client.CheckCarfaxRecords(context.Background(), "1HGCM82633A004352")
	if !IsNotFound(err) {
		t.Fatalf("expected not found error, got %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Message != "record not found" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckCarfaxRecordsMissingVIN(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"vin":"","model":"","carfax_reports":0}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "name", "secret", server.Client())
	_, err := client.CheckCarfaxRecords(context.Background(), "ABC123")
	if !errors.Is(err, ErrNoVIN) {
		t.Fatalf("expected ErrNoVIN, got %v", err)
	}
}
