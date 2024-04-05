package pkg

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"text/template"
	"time"

	getter2 "github.com/hashicorp/go-getter/v2"
	"github.com/spf13/afero"
)

var _ Installer = &DownloadInstaller{}
var Fs = afero.NewOsFs()
var Os = runtime.GOOS

func init() {
	for _, g := range getter2.Getters {
		h, ok := g.(*getter2.HttpGetter)
		if !ok {
			continue
		}
		h.HeadFirstTimeout = time.Duration(0)
		h.ReadTimeout = time.Duration(0)
	}
}

type downloadArgument struct {
	Version string
	Os      string
	Arch    string
}

type DownloadInstaller struct {
	downloadUrlTemplate string

	ctx context.Context
}

func (d *DownloadInstaller) Available() bool {
	return true
}

func NewDownloadInstaller(downloadUrlTemplate string, ctx context.Context) (*DownloadInstaller, error) {
	if ctx == nil {
		ctx = context.TODO()
	}
	d := &DownloadInstaller{
		downloadUrlTemplate: downloadUrlTemplate,
		ctx:                 ctx,
	}
	if err := d.validUrlTemplate(downloadUrlTemplate); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DownloadInstaller) Install(version string, dstPath string) error {
	fmt.Printf("Downloading %s\n", d.DownloadUrl(version))
	_, err := getter2.DefaultClient.Get(d.ctx, &getter2.Request{
		Src:             d.DownloadUrl(version),
		Dst:             filepath.Dir(dstPath),
		GetMode:         getter2.ModeAny,
		Copy:            true,
		DisableSymlinks: true,
	})
	if err != nil {
		fmt.Printf("Failed to download %s: %s\n", d.DownloadUrl(version), err.Error())
	}
	return err
}

func (d *DownloadInstaller) DownloadUrl(version string) string {
	arg := downloadArgument{
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: version,
	}
	var buff bytes.Buffer
	tplt, _ := template.New("download").Parse(d.downloadUrlTemplate)
	_ = tplt.Execute(&buff, arg)
	return buff.String()
}

func (d *DownloadInstaller) validUrlTemplate(templateString string) error {
	arg := downloadArgument{
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: "1.0.0",
	}
	var buff bytes.Buffer
	tplt, err := template.New("download").Parse(templateString)
	if err != nil {
		return err
	}
	return tplt.Execute(&buff, arg)
}
