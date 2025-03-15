package types

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type VersionType struct {
	version *semver.Version
	kind    string // "latest", "nightly", or "version"
}

func NewVersion(s string) (*VersionType, error) {
	s = strings.TrimSpace(s)
	if s == "latest" {
		return &VersionType{kind: "latest"}, nil
	}
	if s == "nightly" {
		return &VersionType{kind: "nightly"}, nil
	}

	// Remove the 'v' prefix if present
	versionStr := s
	if strings.HasPrefix(strings.ToLower(s), "v") {
		versionStr = s[1:]
	}

	version, err := semver.NewVersion(versionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %w", err)
	}

	return &VersionType{
		version: version,
		kind:    "version",
	}, nil
}

// The string representation of the version
func (v *VersionType) String() string {
	switch v.kind {
	case "latest":
		return "latest"
	case "nightly":
		return "nightly"
	default:
		return v.version.String()
	}
}

func (v *VersionType) IsLatest() bool {
	return v.kind == "latest"
}

func (v *VersionType) IsNightly() bool {
	return v.kind == "nightly"
}

func (v *VersionType) Version() *semver.Version {
	return v.version
}
