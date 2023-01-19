package util

import (
	"fmt"

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
	nv, err := sv.NewVersion(v.Version)
	if err != nil {
		return "", err
	}
	nv.IncMinor()
	_, err = nv.SetPrerelease("dev")
	if err != nil {
		return "", err
	}
	return nv.String(), nil
}
