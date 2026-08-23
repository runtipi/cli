package utils

import "testing"

func TestValidateDockerComposeVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantError bool
	}{
		{name: "below minimum", version: "2.33.0", wantError: true},
		{name: "minimum", version: "2.33.1", wantError: false},
		{name: "prefixed minimum", version: "v2.33.1", wantError: false},
		{name: "Docker Desktop minimum", version: "v2.33.1-desktop.1", wantError: false},
		{name: "newer minor", version: "2.34.0", wantError: false},
		{name: "newer major", version: "5.1.2", wantError: false},
		{name: "surrounding whitespace", version: "\n2.33.1\n", wantError: false},
		{name: "malformed", version: "unknown", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateDockerComposeVersion(test.version)
			gotError := err != nil
			if gotError != test.wantError {
				t.Fatalf("validateDockerComposeVersion(%q) error = %v, wantError = %t", test.version, err, test.wantError)
			}
		})
	}
}
