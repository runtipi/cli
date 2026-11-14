package types_test

import (
	"strings"
	"testing"

	"github.com/runtipi/cli/internal/types"
)

func TestVersionIdentifiers(t *testing.T) {
	commitVersion := "sha-" + strings.Repeat("a1", 20)

	for _, test := range []struct {
		input string
		want  string
	}{
		{input: commitVersion, want: commitVersion},
		{input: "  " + commitVersion + "\n", want: commitVersion},
		{input: "latest", want: "latest"},
		{input: "nightly", want: "nightly"},
		{input: "prerelease", want: "prerelease"},
		{input: "4.0.0", want: "v4.0.0"},
		{input: "v4.0.0-beta.1", want: "v4.0.0-beta.1"},
	} {
		t.Run(test.input, func(t *testing.T) {
			version, err := types.NewVersion(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if got := version.String(); got != test.want {
				t.Fatalf("String() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRejectMalformedCommitIdentifiers(t *testing.T) {
	for _, input := range []string{
		"sha-",
		"sha-abcdef0",
		"sha-" + strings.Repeat("a", 39),
		"sha-" + strings.Repeat("a", 41),
		"sha-" + strings.Repeat("g", 40),
		"sha-" + strings.Repeat("A", 40),
		"sha-../../somewhere",
	} {
		t.Run(input, func(t *testing.T) {
			if _, err := types.NewVersion(input); err == nil {
				t.Fatalf("NewVersion(%q) accepted an invalid commit identifier", input)
			}
		})
	}
}
