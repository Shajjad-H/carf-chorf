package kernel

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPricesConfigValidate(t *testing.T) {
	valid := []PricePackage{
		{Credits: 1, AmountCents: 1900, Name: "1 Report Credit"},
		{Credits: 2, AmountCents: 3400, Name: "2 Report Credits"},
	}

	tests := []struct {
		name     string
		packages []PricePackage
		wantErr  bool
	}{
		{name: "valid", packages: valid},
		{name: "empty", packages: nil, wantErr: true},
		{name: "zero credits", packages: []PricePackage{{AmountCents: 1900, Name: "Invalid"}}, wantErr: true},
		{name: "zero amount", packages: []PricePackage{{Credits: 1, Name: "Invalid"}}, wantErr: true},
		{name: "missing name", packages: []PricePackage{{Credits: 1, AmountCents: 1900}}, wantErr: true},
		{name: "duplicate credits", packages: []PricePackage{valid[0], valid[0]}, wantErr: true},
		{name: "missing one credit", packages: []PricePackage{valid[1]}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := pricesConfig{Packages: tc.packages}
			err := cfg.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("Validate() unexpectedly succeeded")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}
}

func TestPricesConfigFile(t *testing.T) {
	data, err := os.ReadFile("../../config/prices.yml")
	if err != nil {
		t.Fatalf("read prices config: %v", err)
	}

	var cfg pricesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("parse prices config: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate prices config: %v", err)
	}

	if len(cfg.Packages) != 3 {
		t.Fatalf("packages length = %d, want 3", len(cfg.Packages))
	}

	wantAmounts := map[uint]uint{1: 1900, 2: 3400, 4: 5900}
	for _, pkg := range cfg.Packages {
		if amount, ok := wantAmounts[pkg.Credits]; !ok || pkg.AmountCents != amount {
			t.Fatalf("unexpected package: %+v", pkg)
		}
	}
}
