package security

import (
	"strings"
	"testing"
)

func TestRedactSecrets(t *testing.T) {
	tests := []struct {
		name string
		in   string
		bad  string
	}{
		{
			name: "bearer header",
			in:   "Authorization: Bearer sk-abcdefghijklmnopqrstuvwxyz123456",
			bad:  "sk-abcdefghijklmnopqrstuvwxyz123456",
		},
		{
			name: "json api key",
			in:   `{"error":"bad","api_key":"sk-abcdefghijklmnopqrstuvwxyz123456"}`,
			bad:  "sk-abcdefghijklmnopqrstuvwxyz123456",
		},
		{
			name: "github token",
			in:   "token=ghp_abcdefghijklmnopqrstuvwxyz1234567890",
			bad:  "ghp_abcdefghijklmnopqrstuvwxyz1234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactSecrets(tt.in)
			if !strings.Contains(got, "[REDACTED]") {
				t.Fatalf("RedactSecrets(%q) = %q, want redaction", tt.in, got)
			}
			if strings.Contains(got, tt.bad) {
				t.Fatalf("RedactSecrets(%q) = %q, still contains %q", tt.in, got, tt.bad)
			}
		})
	}
}
