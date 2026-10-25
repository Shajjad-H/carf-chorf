package money

import "testing"

func TestFormatCents(t *testing.T) {
	tests := map[uint]string{
		0:    "$0.00",
		1900: "$19.00",
		5999: "$59.99",
	}

	for cents, want := range tests {
		if got := FormatCents(cents); got != want {
			t.Errorf("FormatCents(%d) = %q, want %q", cents, got, want)
		}
	}
}
