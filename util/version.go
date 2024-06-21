package util

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/mod/semver"
)

type Version struct {
	Version string
}

func NewVersion(version string) (*Version, error) {
	if !semver.IsValid(version) {
		return nil, fmt.Errorf("%s is invalid", version)
	}
	return &Version{Version: version}, nil
}

func (v Version) Compare(other *Version) int {
	return semver.Compare(v.Version, other.Version)
}

func (v Version) MajorMinor() string {
	return semver.MajorMinor(v.Version)
}

func (v Version) Prerelease() string {
	return semver.Prerelease(v.Version)
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

func (v Version) Minor() string {
	return strings.TrimPrefix(v.MajorMinor(), v.Major())[1:]
}

func (v Version) Major() string {
	return semver.Major(v.Version)
}

func (v Version) IsPatch() bool {
	return v.Patch() != "0"
}

func (v Version) NextMajorMinor() string {
	minor, _ := strconv.Atoi(v.Minor())
	return fmt.Sprintf("%s.%d", v.Major(), minor+1)
}
