package vsphere

import (
	"testing"
)

func TestSanitizeServer(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "bare hostname", in: "vcenter.example.com", want: "vcenter.example.com"},
		{name: "hostname with port", in: "vcenter.example.com:443", want: "vcenter.example.com:443"},
		{name: "https prefix", in: "https://vcenter.example.com", want: "vcenter.example.com"},
		{name: "https with path", in: "https://vcenter.example.com/ui/", want: "vcenter.example.com"},
		{name: "https with sdk path", in: "https://vcenter.example.com/sdk", want: "vcenter.example.com"},
		{name: "http prefix", in: "http://vcenter.example.com", want: "vcenter.example.com"},
		{name: "uppercase scheme", in: "HTTPS://vcenter.example.com/ui", want: "vcenter.example.com"},
		{name: "leading whitespace", in: "  vcenter.example.com  ", want: "vcenter.example.com"},
		{name: "trailing slash only", in: "vcenter.example.com/", want: "vcenter.example.com"},
		{name: "empty string", in: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeServer(tt.in)
			if got != tt.want {
				t.Fatalf("SanitizeServer(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
