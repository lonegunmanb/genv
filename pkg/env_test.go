//go:generate mockgen -destination installer_mock_test.go -package pkg_test . Installer
package pkg_test

import (
	"encoding/json"
	"fmt"
	"github.com/lonegunmanb/genv/pkg"
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"os"
	"path/filepath"
	"testing"
)

type envSuite struct {
	suite.Suite
	stub          *gostub.Stubs
	mockFs        afero.Fs
	mockCtrl      *gomock.Controller
	mockInstaller pkg.Installer
}

func TestEnv(t *testing.T) {
	suite.Run(t, new(envSuite))
}

func (d *envSuite) SetupTest() {
	d.mockFs = afero.NewMemMapFs()
	d.stub = gostub.Stub(&pkg.Fs, d.mockFs).
		Stub(&pkg.Os, "linux")
	d.mockCtrl = gomock.NewController(d.T())
	d.mockInstaller = NewMockInstaller(d.mockCtrl)
}

func (d *envSuite) SetupSubTest() {
	d.SetupTest()
}

func (d *envSuite) TearDownTest() {
	d.stub.Reset()
	d.mockCtrl.Finish()
}

func (d *envSuite) TearDownSubTest() {
	d.TearDownTest()
}

func (d *envSuite) files(fs map[string][]byte) {
	for k, v := range fs {
		_ = afero.WriteFile(d.mockFs, k, v, os.FileMode(0644))
	}
}

func (d *envSuite) TestEnv_Installed() {
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
			sut := pkg.NewEnv("~", "tfenv", "terraform", nil)
			actual, _ := sut.Installed(ver)
			d.Equal(cc.expected, actual)
		})
	}
}

func (d *envSuite) TestInstalled_Windows() {
	stub := gostub.Stub(&pkg.Os, "windows")
	defer stub.Reset()
	d.files(map[string][]byte{
		filepath.Join("/tmp", "tfenv", "v1.0.0", "terraform.exe"): []byte("fake"),
	})
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	d.True(sut.Installed("v1.0.0"))
}

func (d *envSuite) TestUseExistedVersionShouldWriteProfileFile() {
	version := "v1.0.0"
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	profilePath := "/tmp/tfenv/.profile.json"
	d.files(map[string][]byte{
		profilePath: {},
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

func (d *envSuite) TestUseEmptyVersionShouldRemoveVersionFromProfile() {
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", d.mockInstaller)
	profilePath := "/tmp/tfenv/.profile.json"
	d.files(map[string][]byte{
		profilePath: []byte(`{"version":"v1.0.0"}`),
	})
	err := sut.Use("")
	d.NoError(err)
	file, err := afero.ReadFile(d.mockFs, profilePath)
	d.NoError(err)
	var profile pkg.Profile
	err = json.Unmarshal(file, &profile)
	d.NoError(err)
	d.Nil(profile.Version)
}

func (d *envSuite) TestGetCurrentAfterUseShouldReturnUsedVersion() {
	version := "v1.0.0"
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	d.files(map[string][]byte{
		fmt.Sprintf("/tmp/tfenv/%s/terraform", version): []byte("fake"),
	})
	err := sut.Use(version)
	d.NoError(err)
	currentVersion, err := sut.CurrentVersion()
	d.NoError(err)
	d.Equal(version, *currentVersion)
}

func (d *envSuite) TestGetCurrentBeforeUseShouldReturnNil() {
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	currentVersion, err := sut.CurrentVersion()
	d.NoError(err)
	d.Nil(currentVersion)
}

func (d *envSuite) TestCurrentBinaryPath_Installed() {
	version := "v1.0.0"
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	d.files(map[string][]byte{
		fmt.Sprintf("/tmp/tfenv/%s/terraform", version): []byte("fake"),
	})
	err := sut.Use(version)
	d.NoError(err)
	actual, err := sut.CurrentBinaryPath()
	d.NoError(err)
	d.NotNil(actual)
	d.Equal(filepath.Join(string(filepath.Separator), "tmp", "tfenv", version, "terraform"), *actual)
}

func (d *envSuite) TestCurrentBinaryPath_Installed_Windows() {
	stub := gostub.Stub(&pkg.Os, "windows")
	defer stub.Reset()
	version := "v1.0.0"
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	d.files(map[string][]byte{
		fmt.Sprintf("/tmp/tfenv/%s/terraform.exe", version): []byte("fake"),
	})
	err := sut.Use(version)
	d.NoError(err)
	actual, err := sut.CurrentBinaryPath()
	d.NoError(err)
	d.NotNil(actual)
	d.Equal(filepath.Join(string(filepath.Separator), "tmp", "tfenv", version, "terraform.exe"), *actual)
}

func (d *envSuite) TestCurrentBinaryPath_NotInstalled() {
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	actual, err := sut.CurrentBinaryPath()
	d.NoError(err)
	d.Nil(actual)
}

func (d *envSuite) TestUninstall_Installed() {
	version := "v1.0.0"
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	binaryPath := fmt.Sprintf("/tmp/tfenv/%s/terraform", version)
	d.files(map[string][]byte{
		binaryPath: []byte("fake"),
	})
	err := sut.Uninstall(version)
	d.NoError(err)
	exists, err := afero.Exists(d.mockFs, "/tmp/tfenv/v1.0.0/terraform")
	require.NoError(d.T(), err)
	d.False(exists)
	exists, err = afero.Exists(d.mockFs, "/tmp/tfenv/v1.0.0")
	require.NoError(d.T(), err)
	d.False(exists)
}

func (d *envSuite) TestUninstall_NotInstalled() {
	version := "v1.0.0"
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
	err := sut.Uninstall(version)
	d.NoError(err)
	exists, err := afero.Exists(d.mockFs, "/tmp/tfenv/v1.0.0/terraform")
	require.NoError(d.T(), err)
	d.False(exists)
	exists, err = afero.Exists(d.mockFs, "/tmp/tfenv/v1.0.0")
	require.NoError(d.T(), err)
	d.False(exists)
}

func (d *envSuite) TestInstall_AlreadyInstalled() {
	version := "v1.0.0"
	sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
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

func (d *envSuite) TestListInstalled() {
	cases := []struct {
		desc     string
		files    map[string][]byte
		expected []string
	}{
		{
			desc: "tags_installed",
			files: map[string][]byte{
				"/tmp/tfenv/v1.0.0/terraform": []byte("fake"),
				"/tmp/tfenv/v1.1.0/terraform": []byte("fake"),
			},
			expected: []string{
				"v1.0.0",
				"v1.1.0",
			},
		},
		{
			desc: "git_hash_installed",
			files: map[string][]byte{
				"/tmp/tfenv/6eb8fcfb3d7bf1ade345b7d1bd432dadab8f8ad1/terraform": []byte("fake"),
				"/tmp/tfenv/c05e704f072ce244170d80b0d7abb09c86def826/terraform": []byte("fake"),
			},
			expected: []string{
				"6eb8fcfb3d7bf1ade345b7d1bd432dadab8f8ad1",
				"c05e704f072ce244170d80b0d7abb09c86def826",
			},
		},
		{
			desc: "hybrid_installed",
			files: map[string][]byte{
				"/tmp/tfenv/v1.0.0/terraform":                                   []byte("fake"),
				"/tmp/tfenv/c05e704f072ce244170d80b0d7abb09c86def826/terraform": []byte("fake"),
			},
			expected: []string{
				"c05e704f072ce244170d80b0d7abb09c86def826",
				"v1.0.0",
			},
		},
		{
			desc:     "not_installed",
			files:    map[string][]byte{},
			expected: nil,
		},
	}
	for _, c := range cases {
		cc := c
		d.Run(cc.desc, func() {
			d.files(cc.files)
			sut := pkg.NewEnv("/tmp", "tfenv", "terraform", nil)
			installed, err := sut.ListInstalled()
			d.NoError(err)
			d.Equal(cc.expected, installed)
		})
	}
}
