package types

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type VersionType struct {
	version *semver.Version
	kind    string // "latest", "nightly", "prerelease", or "version"
}

func NewVersion(s string) (*VersionType, error) {
	s = strings.TrimSpace(s)

	switch s {
	case "latest":
		return &VersionType{kind: "latest"}, nil
	case "nightly":
		return &VersionType{kind: "nightly"}, nil
	case "prerelease":
		return &VersionType{kind: "prerelease"}, nil
	default:
		version, err := semver.NewVersion(s)
		if err != nil {
			return nil, fmt.Errorf("invalid version format: %w", err)
		}
		return &VersionType{
			version: version,
			kind:    "version",
		}, nil
	}
}

// The string representation of the version
func (v *VersionType) String() string {
	switch v.kind {
	case "latest":
		return "latest"
	case "nightly":
		return "nightly"
	case "prerelease":
		return "prerelease"
	default:
		return fmt.Sprintf("v%s", v.version.String())
	}
}

func (v *VersionType) IsLatest() bool {
	return v.kind == "latest"
}

func (v *VersionType) IsNightly() bool {
	return v.kind == "nightly"
}

func (v *VersionType) IsPrerelease() bool {
	return v.kind == "prerelease"
}

func (v *VersionType) Version() *semver.Version {
	return v.version
}
