package commands_test

import (
	"encoding/json"
	"testing"

	"github.com/runtipi/cli/internal/commands"
)

func TestFindAvailableUpdates(t *testing.T) {
	tests := []struct {
		name           string
		payload        string
		wantUpdates    int
		wantError      bool
		wantTipi       int
		wantLatestTipi int
	}{
		{
			name:           "finds an update",
			payload:        `{"installed":[{"info":{"urn":"bitcoin","version":"1.0.0"},"app":{"version":1},"metadata":{"latestVersion":2,"latestDockerVersion":"2.0.0"}}]}`,
			wantUpdates:    1,
			wantTipi:       1,
			wantLatestTipi: 2,
		},
		{
			name:        "does not report an up to date app",
			payload:     `{"installed":[{"info":{"urn":"bitcoin","version":"2.0.0"},"app":{"version":2},"metadata":{"latestVersion":2,"latestDockerVersion":"2.0.0"}}]}`,
			wantUpdates: 0,
		},
		{
			name:      "rejects a missing installed field",
			payload:   `{}`,
			wantError: true,
		},
		{
			name:      "rejects a null installed field",
			payload:   `{"installed":null}`,
			wantError: true,
		},
		{
			name:      "rejects an incomplete app",
			payload:   `{"installed":[{"info":{"urn":"bitcoin"},"app":{"version":1},"metadata":{"latestVersion":2,"latestDockerVersion":"2.0.0"}}]}`,
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var response commands.InstalledAppsResponse
			if err := json.Unmarshal([]byte(test.payload), &response); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}

			updates, err := commands.FindAvailableUpdates(response)
			if (err != nil) != test.wantError {
				t.Fatalf("FindAvailableUpdates() error = %v, want error: %v", err, test.wantError)
			}
			if err != nil {
				return
			}
			if len(updates) != test.wantUpdates {
				t.Fatalf("FindAvailableUpdates() returned %d updates, want %d", len(updates), test.wantUpdates)
			}
			if len(updates) == 1 {
				if updates[0].TipiVersion != test.wantTipi || updates[0].LatestTipi != test.wantLatestTipi {
					t.Fatalf("FindAvailableUpdates() returned Tipi versions %d -> %d, want %d -> %d", updates[0].TipiVersion, updates[0].LatestTipi, test.wantTipi, test.wantLatestTipi)
				}
			}
		})
	}
}
