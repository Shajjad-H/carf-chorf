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

func TestPayPalConfigValidate(t *testing.T) {
	valid := paypalConfig{
		ClientID:     "client",
		ClientSecret: "secret",
		APIURL:       "https://api-m.sandbox.paypal.com",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name string
		cfg  paypalConfig
	}{
		{name: "missing client id", cfg: paypalConfig{ClientSecret: valid.ClientSecret, APIURL: valid.APIURL}},
		{name: "missing secret", cfg: paypalConfig{ClientID: valid.ClientID, APIURL: valid.APIURL}},
		{name: "invalid url", cfg: paypalConfig{ClientID: valid.ClientID, ClientSecret: valid.ClientSecret, APIURL: "not-a-url"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err == nil {
				t.Fatal("Validate() unexpectedly succeeded")
			}
		})
	}
}

func TestDiscordConfigValidate(t *testing.T) {
	valid := discordConfig{
		ContactWebhookURL: "https://discord.com/api/webhooks/contact",
		SOSWebhookURL:     "https://discord.com/api/webhooks/sos",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	tests := []struct {
		name string
		cfg  discordConfig
	}{
		{name: "missing contact", cfg: discordConfig{SOSWebhookURL: valid.SOSWebhookURL}},
		{name: "missing sos", cfg: discordConfig{ContactWebhookURL: valid.ContactWebhookURL}},
		{
			name: "invalid sos",
			cfg: discordConfig{
				ContactWebhookURL: valid.ContactWebhookURL,
				SOSWebhookURL:     "not-a-url",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err == nil {
				t.Fatal("Validate() unexpectedly succeeded")
			}
		})
	}
}

func TestDashboardConfigValidate(t *testing.T) {
	valid := dashboardConfig{
		Address:  "localhost:8080",
		AdminKey: "12345678901234567890",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	longer := valid
	longer.AdminKey = "1234567890123456789012345"
	if err := longer.Validate(); err != nil {
		t.Fatalf("longer admin key rejected: %v", err)
	}

	tests := []struct {
		name string
		cfg  dashboardConfig
	}{
		{name: "missing address", cfg: dashboardConfig{AdminKey: valid.AdminKey}},
		{name: "missing key", cfg: dashboardConfig{Address: valid.Address}},
		{name: "too short", cfg: dashboardConfig{Address: valid.Address, AdminKey: "1234567890123456789"}},
		{name: "non numeric", cfg: dashboardConfig{Address: valid.Address, AdminKey: "1234567890123456789a"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.cfg.Validate(); err == nil {
				t.Fatal("Validate() unexpectedly succeeded")
			}
		})
	}
}
