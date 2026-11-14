package utils_test

import (
	"strings"
	"testing"

	"github.com/runtipi/cli/internal/utils"
)

func TestBranchBuildsAreExplicitUpdates(t *testing.T) {
	commitVersion := "sha-" + strings.Repeat("a1", 20)
	releases := []utils.GitHubRelease{
		{TagName: commitVersion, Prerelease: true},
		{TagName: "nightly", Prerelease: true},
		{TagName: "e2e", Prerelease: true},
		{TagName: "v4.0.0", Prerelease: false},
		{TagName: "v4.1.0-beta.1", Prerelease: true},
	}

	filtered := utils.FilterNonPreReleases(releases)
	if len(filtered) != 1 || filtered[0].TagName != "v4.1.0-beta.1" {
		t.Fatalf("prerelease selection includes an unrelated build: %#v", filtered)
	}
	if utils.IsMajorBump("v4.0.0", commitVersion) {
		t.Fatal("commit identifiers must not be interpreted as semantic versions")
	}
	if !utils.IsMajorBump("v3.0.0", "v4.0.0") {
		t.Fatal("regular major-version upgrades must remain guarded")
	}
}
