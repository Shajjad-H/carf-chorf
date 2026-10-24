package contact

import (
	"strings"
	"testing"
)

func TestContactRequestValidate(t *testing.T) {
	longMessage := strings.Repeat("a", 2001)

	tests := []struct {
		name      string
		req       ContactRequest
		wantErr   bool
		wantPhone string
	}{
		{
			name:      "valid",
			req:       ContactRequest{Email: "user@example.com", Phone: "(555) 123-4567", Message: "Hello, I need help with my order."},
			wantPhone: "+1 (555) 123-4567",
		},
		{
			name:      "valid with country code",
			req:       ContactRequest{Email: "user@example.com", Phone: "+1 555-123-4567", Message: "Hello, I need help with my order."},
			wantPhone: "+1 (555) 123-4567",
		},
		{
			name:    "invalid email",
			req:     ContactRequest{Email: "nope", Phone: "555-123-4567", Message: "Hello, I need help."},
			wantErr: true,
		},
		{
			name:    "invalid phone",
			req:     ContactRequest{Email: "user@example.com", Phone: "12345", Message: "Hello, I need help."},
			wantErr: true,
		},
		{
			name:    "message too short",
			req:     ContactRequest{Email: "user@example.com", Phone: "555-123-4567", Message: "short"},
			wantErr: true,
		},
		{
			name:    "message too long",
			req:     ContactRequest{Email: "user@example.com", Phone: "555-123-4567", Message: longMessage},
			wantErr: true,
		},
		{
			name:    "missing fields",
			req:     ContactRequest{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.req.Phone != tt.wantPhone {
				t.Fatalf("phone = %q, want %q", tt.req.Phone, tt.wantPhone)
			}
		})
	}
}
