package main

import (
	"os"
	"testing"
)

func TestResolveAddr(t *testing.T) {
	tests := []struct {
		name     string
		flagPort int
		envPort  string
		want     string
	}{
		{
			name:     "flag takes precedence",
			flagPort: 2346,
			envPort:  "9999",
			want:     "localhost:2346",
		},
		{
			name:     "env var used when flag is zero",
			flagPort: 0,
			envPort:  "2347",
			want:     "localhost:2347",
		},
		{
			name:     "default when no flag or env",
			flagPort: 0,
			envPort:  "",
			want:     "localhost:2345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envPort != "" {
				os.Setenv("DLV_PORT", tt.envPort)
				defer os.Unsetenv("DLV_PORT")
			} else {
				os.Unsetenv("DLV_PORT")
			}

			got := resolveAddr(tt.flagPort)
			if got != tt.want {
				t.Errorf("resolveAddr(%d) = %q, want %q", tt.flagPort, got, tt.want)
			}
		})
	}
}
