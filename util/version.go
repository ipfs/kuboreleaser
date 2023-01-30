package util

import (
	"fmt"
	"strings"

	sv "github.com/Masterminds/semver"
	"golang.org/x/mod/semver"
)

type Version struct {
	Version string
}

func NewVersion(version string) (*Version, error) {
	if !semver.IsValid(version) {
		return nil, fmt.Errorf("invalid version: %s", version)
	}
	return &Version{Version: version}, nil
}

func (v Version) MajorMinor() string {
	return semver.MajorMinor(v.Version)
}

func (v Version) Prerelease() string {
	return semver.Prerelease(v.Version)
}

func (v Version) Dev() (string, error) {
	newVersion, err := sv.NewVersion(v.Version)
	if err != nil {
		return "", err
	}
	nextVersion := newVersion.IncMinor()
	devVersion, err := nextVersion.SetPrerelease("dev")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("v%s", devVersion.String()), nil
}

func (v Version) IsPrerelease() bool {
	return v.Prerelease() != ""
}

func (v Version) String() string {
	return v.Version
}

func (v Version) MajorMinorPatch() string {
	return strings.TrimSuffix(v.Version, semver.Prerelease(v.Version)+semver.Build(v.Version))
}

func (v Version) Patch() string {
	return strings.TrimPrefix(v.MajorMinorPatch(), v.MajorMinor())[1:]
}

func (v Version) IsPatch() bool {
	return v.Patch() != "0"
}
