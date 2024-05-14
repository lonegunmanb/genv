package pkg_test

import (
	"context"
	"github.com/lonegunmanb/genv/pkg"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
)

type goBuildInstallerSuite struct {
	suite.Suite
	outputFolder string
}

func TestGoBuildInstaller(t *testing.T) {
	suite.Run(t, new(goBuildInstallerSuite))
}

func (g *goBuildInstallerSuite) SetupTest() {
	var err error
	g.outputFolder, err = os.MkdirTemp("", "")
	if err != nil {
		panic(err.Error())
	}
}

func (g *goBuildInstallerSuite) SetupSubTest() {
	g.SetupTest()
}

func (g *goBuildInstallerSuite) TearDownTest() {
	_ = os.RemoveAll(g.outputFolder)
}

func (g *goBuildInstallerSuite) TearDownSubTest() {
	g.TearDownTest()
}

func (g *goBuildInstallerSuite) TestInstall() {
	cases := []struct {
		desc    string
		version string
		binary  string
		repoUrl string
	}{
		{
			desc:    "tag",
			version: "v1.0.0",
			binary:  "http-echo",
			repoUrl: "https://github.com/hashicorp/http-echo.git",
		},
		{
			desc:    "latest",
			version: "latest",
			binary:  "http-echo",
			repoUrl: "https://github.com/hashicorp/http-echo.git",
		},
		{
			desc:    "git_hash",
			version: "eed3dec30a2cc2e79d1fff131fa7025c52950c12",
			binary:  "http-echo",
			repoUrl: "https://github.com/hashicorp/http-echo.git",
		},
		{
			desc:    "different_binary_name",
			version: "latest",
			binary:  "http-echo-custom",
			repoUrl: "https://github.com/hashicorp/http-echo.git",
		},
		{
			desc:    "sub_path_without_gomod",
			version: "975961f5f06346ccc282cd0d9aa16e160d26f9e3",
			binary:  "go-getter",
			repoUrl: "https://github.com/hashicorp/go-getter.git",
		},
		{
			desc:    "sub_path_wit_gomod",
			version: "v2.2.0",
			binary:  "go-getter",
			repoUrl: "https://github.com/hashicorp/go-getter.git",
		},
	}
	for _, c := range cases {
		cc := c
		g.Run(cc.desc, func() {
			binary := cc.binary
			if runtime.GOOS == "windows" {
				binary += ".exe"
			}
			installer := pkg.NewGoBuildInstaller(cc.repoUrl, binary, "", context.Background())
			dstPath := filepath.Join(g.outputFolder, binary)
			err := installer.Install(cc.version, dstPath)
			g.NoError(err)
			exist, err := fileExist(dstPath)
			g.NoError(err)
			g.True(exist)
		})
	}
}

func fileExist(path string) (found bool, err error) {
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
	} else {
		found = true
	}

	return
}
