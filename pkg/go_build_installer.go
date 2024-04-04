package pkg

import (
	"context"
	"crypto/rand"
	"fmt"
	getter2 "github.com/hashicorp/go-getter/v2"
	"os"
	"os/exec"
	"path/filepath"
)

var _ Installer = &GoBuildInstaller{}

type GoBuildInstaller struct {
	repoUrl    string
	subPath    string
	ctx        context.Context
	binaryName string
}

func NewGoBuildInstaller(repoUrl string, binaryName string, subPath string, ctx context.Context) *GoBuildInstaller {
	if ctx == nil {
		ctx = context.TODO()
	}
	return &GoBuildInstaller{
		repoUrl:    repoUrl,
		subPath:    subPath,
		ctx:        ctx,
		binaryName: binaryName,
	}
}

func (g *GoBuildInstaller) Install(version string, dstPath string) error {
	tmpDir := filepath.Join(os.TempDir(), randStr(8))
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	src := g.repoUrl
	if version != "latest" {
		src = fmt.Sprintf("%s?ref=%s", g.repoUrl, version)
	}
	src = fmt.Sprintf("git::%s", src)
	_, err := getter2.Get(g.ctx, tmpDir, src)
	if err != nil {
		return err
	}
	err = executeCommand(tmpDir, "go", "mod", "download")
	if err != nil {
		return err
	}
	args := []string{"build", "-o", fmt.Sprintf(filepath.Join(dstPath, g.binaryName))}
	if g.subPath != "" {
		args = append(args, g.subPath)
	}
	return executeCommand(tmpDir, "go", args...)
}

func (g *GoBuildInstaller) Available() bool {
	return exec.Command("go", "version").Run() == nil
}

func executeCommand(wd string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = wd
	return cmd.Run()
}

func randStr(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
