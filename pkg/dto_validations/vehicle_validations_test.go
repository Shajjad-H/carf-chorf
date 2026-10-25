package dto_validations

import "testing"

func TestLookupRequestValidateVIN(t *testing.T) {
	req := LookupRequest{Vin: " 1hgcm82633a004352 ", Plate: "ABC123", State: "CA"}
	if err := req.Validate(); err != nil {
		t.Fatalf("valid VIN rejected: %v", err)
	}
	if req.Vin != "1HGCM82633A004352" {
		t.Fatalf("normalization failed: %+v", req)
	}
	if req.Plate != "" || req.State != "" {
		t.Fatalf("VIN lookup kept plate/state: %+v", req)
	}

	for _, vin := range []string{"SHORT", "1HGCM82633A00435I", "1HGCM82633A00435Q"} {
		req := LookupRequest{Vin: vin, Plate: "ABC123", State: "CA"}
		if err := req.Validate(); err == nil {
			t.Errorf("invalid VIN %q accepted over valid plate/state", vin)
		}
	}
}

func TestLookupRequestValidatePlate(t *testing.T) {
	req := LookupRequest{Plate: " abc-123 ", State: " ca "}
	if err := req.Validate(); err != nil {
		t.Fatalf("valid plate rejected: %v", err)
	}
	if req.Plate != "ABC-123" || req.State != "CA" {
		t.Fatalf("normalization failed: %+v", req)
	}
	if req.Vin != "" {
		t.Fatalf("plate lookup kept VIN: %+v", req)
	}

	invalid := []LookupRequest{
		{Plate: "", State: "CA"},
		{Plate: "ABC123", State: ""},
		{Plate: "ABC123", State: "C"},
		{Plate: "TOO-LONG-PLATE", State: "CA"},
	}
	for _, req := range invalid {
		if err := req.Validate(); err == nil {
			t.Errorf("invalid plate request accepted: %+v", req)
		}
	}
}

func TestLookupRequestEmpty(t *testing.T) {
	req := LookupRequest{}
	if err := req.Validate(); err == nil {
		t.Fatal("empty lookup request accepted")
	}
}
