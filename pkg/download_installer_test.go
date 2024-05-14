package pkg_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lonegunmanb/genv/pkg"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type downloadInstallerSuite struct {
	suite.Suite
	stub   *gostub.Stubs
	mockFs afero.Fs
}

func TestDownloadInstaller(t *testing.T) {
	suite.Run(t, new(downloadInstallerSuite))
}

func (d *downloadInstallerSuite) SetupTest() {
	d.mockFs = afero.NewMemMapFs()
	d.stub = gostub.Stub(&pkg.Fs, d.mockFs).
		Stub(&pkg.Os, "linux")
}

func (d *downloadInstallerSuite) SetupSubTest() {
	d.SetupTest()
}

func (d *downloadInstallerSuite) TearDownTest() {
	d.stub.Reset()
}

func (d *downloadInstallerSuite) TearDownSubTest() {
	d.TearDownTest()
}

func (d *downloadInstallerSuite) TestInstall_DownloadUrl() {
	version := "1.7.5"
	sut, _ := pkg.NewDownloadInstaller("https://releases.hashicorp.com/terraform/{{ .Version }}/terraform_{{ .Version }}_{{ .Os }}_{{ .Arch }}.zip", nil)
	d.Equal(fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip", version, version, runtime.GOOS, runtime.GOARCH), sut.DownloadUrl(version))
}

func (d *downloadInstallerSuite) TestIncorrectDownloadUrlTemplateShouldReturnError() {
	incorrectCases := []struct {
		desc string
		url  string
	}{
		{
			desc: "no-supported argument",
			url:  "https://releases.hashicorp.com/terraform//terraform_{{ .Unknown }}.zip",
		},
	}
	for _, c := range incorrectCases {
		cc := c
		d.Run(cc.desc, func() {
			_, err := pkg.NewDownloadInstaller(cc.url, nil)
			d.NotNil(err)
		})
	}

}

func TestInstall_Install(t *testing.T) {
	version := "1.7.5"
	sut, _ := pkg.NewDownloadInstaller("https://releases.hashicorp.com/terraform/{{ .Version }}/terraform_{{ .Version }}_{{ .Os }}_{{ .Arch }}.zip", context.Background())
	defer func() {
		_ = os.RemoveAll(filepath.Join("/tmp", "tfenv"))
	}()
	binary := "terraform"
	if runtime.GOOS == "windows" {
		binary = "terraform.exe"
	}
	binaryPath := filepath.Join("/tmp", "tfenv", version, binary)
	_ = os.Remove(binaryPath)
	_, err := os.Stat(binaryPath)
	require.True(t, errors.Is(err, os.ErrNotExist))
	err = sut.Install(version, binaryPath)
	require.NoError(t, err)
	stat, err := os.Stat(binaryPath)
	require.NoError(t, err)
	assert.False(t, stat.IsDir())
	cmd := exec.Cmd{
		Path: binaryPath,
		Args: []string{binaryPath, "version", "-json"},
	}
	output, err := cmd.Output()
	require.NoError(t, err)
	var outputMap map[string]any
	err = json.Unmarshal(output, &outputMap)
	require.NoError(t, err)
	assert.Equal(t, version, outputMap["terraform_version"])
}
