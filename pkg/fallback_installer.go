package pkg

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var _ Installer = &fallbackInstaller{}

type fallbackInstaller struct {
	i1 Installer
	i2 Installer
}

func (f *fallbackInstaller) Install(version string, dstPath string) error {
	err := f.install(f.i1, version, dstPath)
	if err != nil {
		return f.install(f.i2, version, dstPath)
	}
	return nil
}

func (f *fallbackInstaller) Available() bool {
	return f.i1.Available() || f.i2.Available()
}

func NewFallbackInstaller(i1 Installer, i2 Installer) Installer {
	return &fallbackInstaller{
		i1: i1,
		i2: i2,
	}
}

func (f *fallbackInstaller) install(i Installer, version string, dstPath string) error {
	_, err := semver.NewVersion(version)
	isSemver := err == nil
	if !isSemver {
		return i.Install(version, dstPath)
	}
	err = i.Install(version, dstPath)
	if err == nil {
		return nil
	}
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	} else {
		version = fmt.Sprintf("v%s", version)
	}
	return i.Install(version, dstPath)
}
