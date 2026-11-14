package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type VersionType struct {
	version *semver.Version
	kind    string
	commit  string
}

var commitVersionPattern = regexp.MustCompile(`^sha-[0-9a-f]{40}$`)

func NewVersion(s string) (*VersionType, error) {
	s = strings.TrimSpace(s)

	if commitVersionPattern.MatchString(s) {
		return &VersionType{kind: "commit", commit: s}, nil
	}

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

func (v *VersionType) String() string {
	switch v.kind {
	case "commit":
		return v.commit
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

func (v *VersionType) IsCommit() bool {
	return v.kind == "commit"
}

func (v *VersionType) Version() *semver.Version {
	return v.version
}
