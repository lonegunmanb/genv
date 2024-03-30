package pkg_test

import (
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
