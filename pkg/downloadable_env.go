package pkg

import (
	"github.com/spf13/afero"
	"path/filepath"
)

var _ Env = DownloadableEnv{}
var Fs = afero.NewOsFs()

type DownloadableEnv struct {
	downloadUrlTemplate string
	homeDir             string
	name                string
	binaryName          string
}

func NewDownloadableEnv(downloadUrlTemplate, homeDir, name, binaryName string) DownloadableEnv {
	return DownloadableEnv{
		downloadUrlTemplate: downloadUrlTemplate,
		homeDir:             homeDir,
		name:                name,
		binaryName:          binaryName,
	}
}

func (d DownloadableEnv) CurrentBinaryPath() (string, error) {
	panic("implement me")
}

func (d DownloadableEnv) Name() string {
	return d.name
}

func (d DownloadableEnv) BinaryName() string {
	return d.binaryName
}

func (d DownloadableEnv) Installed(version string) (bool, error) {
	b, err := afero.Exists(Fs, d.binaryPath(version))
	if err != nil {
		return false, err
	}
	return b, nil
}

func (d DownloadableEnv) Install(version string) error {
	//TODO implement me
	panic("implement me")
}

func (d DownloadableEnv) Use(version string) error {
	//TODO implement me
	panic("implement me")
}

func (d DownloadableEnv) CurrentVersion() *string {
	//TODO implement me
	panic("implement me")
}

func (d DownloadableEnv) binaryPath(version string) string {
	return filepath.Join(d.homeDir, d.name, version, d.binaryName)
}
