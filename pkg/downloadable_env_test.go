package pkg_test

import (
	"encoding/json"
	"fmt"
	"github.com/lonegunmanb/genv/pkg"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"os"
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
			sut := pkg.NewDownloadableEnv("", "~", "tfenv", "terraform")
			actual, _ := sut.Installed(ver)
			d.Equal(cc.expected, actual)
		})
	}
}

func (d *downloadableEnvSuite) TestUseNonExistVersionShouldThrowError() {
	sut := pkg.NewDownloadableEnv("", "~", "tfenv", "terraform")
	err := sut.Use("v1.0.0")
	d.NotNil(err)
	d.Contains(err.Error(), "not installed")
}

func (d *downloadableEnvSuite) TestUseExistedVersionShouldWriteProfileFile() {
	version := "v1.0.0"
	sut := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform")
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
	sut := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform")
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
	sut := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform")
	currentVersion, err := sut.CurrentVersion()
	d.NoError(err)
	d.Nil(currentVersion)
}

func (d *downloadableEnvSuite) TestCurrentBinaryPath_Installed() {
	version := "v1.0.0"
	sut := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform")
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
	sut := pkg.NewDownloadableEnv("", "/tmp", "tfenv", "terraform")
	actual, err := sut.CurrentBinaryPath()
	d.NoError(err)
	d.Nil(actual)
}
