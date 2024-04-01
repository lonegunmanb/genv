package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	getter2 "github.com/hashicorp/go-getter/v2"
	"github.com/xianic/fslock"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/spf13/afero"
)

var _ Env = &DownloadableEnv{}
var Fs = afero.NewOsFs()

type downloadArgument struct {
	Version string
	Os      string
	Arch    string
}

type DownloadableEnv struct {
	downloadUrlTemplate string
	homeDir             string
	name                string
	binaryName          string
	l                   *fslock.Lock
	ctx                 context.Context
}

func NewDownloadableEnv(downloadUrlTemplate, homeDir, name, binaryName string, ctx context.Context) (*DownloadableEnv, error) {
	if ctx == nil {
		ctx = context.TODO()
	}
	d := &DownloadableEnv{
		downloadUrlTemplate: downloadUrlTemplate,
		homeDir:             homeDir,
		name:                name,
		binaryName:          binaryName,
		ctx:                 ctx,
	}
	if err := d.validUrlTemplate(downloadUrlTemplate); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DownloadableEnv) CurrentBinaryPath() (*string, error) {
	ver, err := d.CurrentVersion()
	if err != nil {
		return nil, err
	}
	if ver == nil {
		return nil, nil
	}

	p := filepath.Join(d.homeDir, d.name, *ver, d.binaryName)
	return &p, nil
}

func (d *DownloadableEnv) Name() string {
	return d.name
}

func (d *DownloadableEnv) BinaryName() string {
	return d.binaryName
}

func (d *DownloadableEnv) Installed(version string) (bool, error) {
	b, err := afero.Exists(Fs, d.binaryPath(version))
	if err != nil {
		return false, err
	}
	return b, nil
}

func (d *DownloadableEnv) Install(version string) error {
	installed, err := d.Installed(version)
	if err != nil {
		return err
	}
	if installed {
		return nil
	}
	getter := getter2.DefaultClient
	_, err = getter.Get(d.ctx, &getter2.Request{
		Src:             d.DownloadUrl(version),
		Dst:             filepath.Dir(d.binaryPath(version)),
		GetMode:         getter2.ModeAny,
		Copy:            true,
		DisableSymlinks: true,
	})
	return err
}

func (d *DownloadableEnv) Use(version string) error {
	installed, err := d.Installed(version)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("version %s not installed, please install it first", version)
	}
	err = d.lock()
	if err != nil {
		return err
	}
	defer func() {
		_ = d.unlock()
	}()
	profilePath := d.profilePath()
	profile, err := d.profile()
	if err != nil {
		return err
	}
	if profile == nil {
		profile = &Profile{}
	}
	profile.Version = &version
	profileContent, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	return afero.WriteFile(Fs, profilePath, profileContent, os.ModePerm)
}

func (d *DownloadableEnv) CurrentVersion() (*string, error) {
	profile, err := d.profile()
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}
	return profile.Version, nil
}

func (d *DownloadableEnv) DownloadUrl(version string) string {
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

func (d *DownloadableEnv) validUrlTemplate(templateString string) error {
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

func (d *DownloadableEnv) binaryPath(version string) string {
	return filepath.Join(d.homeDir, d.name, version, d.binaryName)
}

func (d *DownloadableEnv) profile() (*Profile, error) {
	exist, err := afero.Exists(Fs, d.profilePath())
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	profileContent, err := afero.ReadFile(Fs, d.profilePath())
	if err != nil {
		return nil, err
	}
	var profile Profile
	_ = json.Unmarshal(profileContent, &profile)
	return &profile, nil
}

func (d *DownloadableEnv) profilePath() string {
	return filepath.Join(d.homeDir, d.name, ".profile.json")
}

func (d *DownloadableEnv) lockPath() string {
	return filepath.Join(d.homeDir, d.name, ".lock")
}

func (d *DownloadableEnv) ensureHomeDir() {
	dir := filepath.Join(d.homeDir, d.name)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		_ = os.MkdirAll(dir, 0755)
	}
}

func (d *DownloadableEnv) lock() error {
	if d.l != nil {
		return fmt.Errorf("this env has already been locked")
	}
	lockPath := d.lockPath()
	d.ensureHomeDir()
	lock := fslock.New(lockPath)
	err := lock.Lock()
	if err != nil {
		return err
	}
	d.l = lock
	return nil
}

func (d *DownloadableEnv) unlock() error {
	if d.l == nil {
		return nil
	}
	if err := d.l.Unlock(); err != nil {
		return err
	}
	d.l = nil
	return nil
}
