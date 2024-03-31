package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blend/go-sdk/filelock"
	"github.com/spf13/afero"
)

var _ Env = &DownloadableEnv{}
var Fs = afero.NewOsFs()

type DownloadableEnv struct {
	downloadUrlTemplate string
	homeDir             string
	name                string
	binaryName          string
	lockFile            *os.File
}

func NewDownloadableEnv(downloadUrlTemplate, homeDir, name, binaryName string) *DownloadableEnv {
	return &DownloadableEnv{
		downloadUrlTemplate: downloadUrlTemplate,
		homeDir:             homeDir,
		name:                name,
		binaryName:          binaryName,
	}
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
	//TODO implement me
	panic("implement me")
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
	if d.lockFile != nil {
		return fmt.Errorf("this env has already been locked")
	}
	lockPath := d.lockPath()
	d.ensureHomeDir()
	if _, err := os.Stat(lockPath); errors.Is(err, os.ErrNotExist) {
		f, _ := os.Create(lockPath)
		_ = f.Close()
	}
	f, err := os.OpenFile(lockPath, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	d.lockFile = f
	return filelock.Lock(f)
}

func (d *DownloadableEnv) unlock() error {
	if d.lockFile == nil {
		return nil
	}
	if err := filelock.Unlock(d.lockFile); err != nil {
		return err
	}
	f := d.lockFile
	d.lockFile = nil
	_ = os.Remove(f.Name())
	return nil
}
