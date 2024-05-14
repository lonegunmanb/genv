package pkg

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	getter2 "github.com/hashicorp/go-getter/v2"
)

var _ Installer = &GoBuildInstaller{}

type GoBuildInstaller struct {
	repoUrl    string
	subPath    string
	ctx        context.Context
	binaryName string
}

func NewGoBuildInstaller(repoUrl string, binaryName string, subPath string, ctx context.Context) Installer {
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
	fmt.Printf("Go build %s\n", g.repoUrl)
	_, err := getter2.Get(g.ctx, tmpDir, src)
	if err != nil {
		fmt.Printf("Failed to clone %s: %s\n", g.repoUrl, err.Error())
		return err
	}
	fmt.Printf("go mod download at %s\n", tmpDir)
	err = executeCommand(tmpDir, "go", "mod", "download")
	if err != nil {
		fmt.Printf("Failed to download go mod at %s: %s\n", tmpDir, err.Error())
		return err
	}
	args := []string{"build", "-o", dstPath}
	if g.subPath != "" {
		args = append(args, g.subPath)
	}
	fmt.Printf("go build -o %s\n", args[2])
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
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
