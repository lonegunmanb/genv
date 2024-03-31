package pkg_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lonegunmanb/genv/pkg"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

type downloadableEnvSuite struct {
	suite.Suite
	stub   *gostub.Stubs
	mockFs afero.Fs
}

func TestDownloadableEnv(t *testing.T) {
	suite.Run(t, new(downloadableEnvSuite))
}

func (d *downloadableEnvSuite) SetupTest() {
	d.mockFs = afero.NewMemMapFs()
	d.stub = gostub.Stub(&pkg.Fs, d.mockFs)
}

func (d *downloadableEnvSuite) SetupSubTest() {
	d.SetupTest()
}

func (d *downloadableEnvSuite) TearDownTest() {
	d.stub.Reset()
}

func (d *downloadableEnvSuite) TearDownSubTest() {
	d.TearDownTest()
}

func (d *downloadableEnvSuite) files(fs map[string][]byte) {
	for k, v := range fs {
		_ = afero.WriteFile(d.mockFs, k, v, os.FileMode(644))
	}
}

func (d *downloadableEnvSuite) TestDownloadableEnv_Installed() {
	ver := "v1.0.0"
	cases := []struct {
		desc     string
		fs       map[string][]byte
		expected bool
	}{
		{
			desc: "installed",
			fs: map[string][]byte{
				fmt.Sprintf("~/tfenv/%s/terraform", ver): []byte("fake"),
			},
			expected: true,
		},
		{
			desc: "not installed",
			fs: map[string][]byte{
				"~/tfenv/v1.0.1/terraform": []byte("fake"),
			},
			expected: false,
		},
		{
			desc: "not installed, no version folder",
			fs: map[string][]byte{
				"~/tfenv/.lock": []byte("fake"),
			},
			expected: false,
		},
		{
			desc: "not installed, env",
			fs: map[string][]byte{
				"~/.lock": []byte("fake"),
			},
			expected: false,
		},
	}
	for _, c := range cases {
		cc := c
		d.Run(cc.desc, func() {
			d.files(cc.fs)
			sut, _ := pkg.NewDownloadableEnv("", "~", "tfenv", "terraform", nil)
			actual, _ := sut.Installed(ver)
			d.Equal(cc.expected, actual)
		})
	}
}

func (d *downloadableEnvSuite) TestUseNonExistVersionShouldThrowError() {
	sut, _ := pkg.NewDownloadableEnv("", "~", "tfenv", "terraform", nil)
	err := sut.Use("v1.0.0")
	d.NotNil(err)
	d.Contains(err.Error(), "not installed")
}

func (d *downloadableEnvSuite) TestUseExistedVersionShouldWriteProfileFile() {
	version := "v1.0.0"
	sut, _ := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform", nil)
	profilePath := "/tmp/tfenv/.profile.json"
	d.files(map[string][]byte{
		profilePath: []byte{},
		fmt.Sprintf("/tmp/tfenv/%s/terraform", version): []byte("fake"),
	})
	err := sut.Use(version)
	d.NoError(err)
	file, err := afero.ReadFile(d.mockFs, profilePath)
	d.NoError(err)
	var profile pkg.Profile
	err = json.Unmarshal(file, &profile)
	d.NoError(err)
	d.Equal(version, *profile.Version)
}

func (d *downloadableEnvSuite) TestGetCurrentAfterUseShouldReturnUsedVersion() {
	version := "v1.0.0"
	sut, _ := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform", nil)
	d.files(map[string][]byte{
		fmt.Sprintf("/tmp/tfenv/%s/terraform", version): []byte("fake"),
	})
	err := sut.Use(version)
	d.NoError(err)
	currentVersion, err := sut.CurrentVersion()
	d.NoError(err)
	d.Equal(version, *currentVersion)
}

func (d *downloadableEnvSuite) TestGetCurrentBeforeUseShouldReturnNil() {
	sut, _ := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform", nil)
	currentVersion, err := sut.CurrentVersion()
	d.NoError(err)
	d.Nil(currentVersion)
}

func (d *downloadableEnvSuite) TestCurrentBinaryPath_Installed() {
	version := "v1.0.0"
	sut, _ := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform", nil)
	d.files(map[string][]byte{
		fmt.Sprintf("/tmp/tfenv/%s/terraform", version): []byte("fake"),
	})
	err := sut.Use(version)
	d.NoError(err)
	actual, err := sut.CurrentBinaryPath()
	d.NoError(err)
	d.NotNil(actual)
	d.Equal("/tmp/tfenv/v1.0.0/terraform", *actual)
}

func (d *downloadableEnvSuite) TestCurrentBinaryPath_NotInstalled() {
	sut, _ := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform", nil)
	actual, err := sut.CurrentBinaryPath()
	d.NoError(err)
	d.Nil(actual)
}

func (d *downloadableEnvSuite) TestInstall_AlreadyInstalled() {
	version := "v1.0.0"
	sut, _ := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform", nil)
	binaryPath := fmt.Sprintf("/tmp/tfenv/%s/terraform", version)
	d.files(map[string][]byte{
		binaryPath: []byte("fake"),
	})
	err := sut.Install(version)
	d.NoError(err)
	file, err := afero.ReadFile(d.mockFs, binaryPath)
	d.NoError(err)
	d.Equal("fake", string(file))
}

func (d *downloadableEnvSuite) TestInstall_DownloadUrl() {
	version := "1.7.5"
	sut, _ := pkg.NewDownloadableEnv("https://releases.hashicorp.com/terraform/{{ .Version }}/terraform_{{ .Version }}_{{ .Os }}_{{ .Arch }}.zip", "/tmp", "tfenv", "terraform", nil)
	d.Equal(fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip", version, version, runtime.GOOS, runtime.GOARCH), sut.DownloadUrl(version))
}

func (d *downloadableEnvSuite) TestIncorrectDownloadUrlTemplateShouldReturnError() {
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
			_, err := pkg.NewDownloadableEnv(cc.url, "/tmp", "tfenv", "terraform", nil)
			d.NotNil(err)
		})
	}

}

func TestInstall_Install(t *testing.T) {
	version := "1.7.5"
	sut, _ := pkg.NewDownloadableEnv("https://releases.hashicorp.com/terraform/{{ .Version }}/terraform_{{ .Version }}_{{ .Os }}_{{ .Arch }}.zip", "/tmp", "tfenv", "terraform", nil)
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
	err = sut.Install(version)
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
