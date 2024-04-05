package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/xianic/fslock"
)

type Env struct {
	homeDir    string
	name       string
	binaryName string
	l          *fslock.Lock
	Installer
}

func NewEnv(homeDir, name, binaryName string, installer Installer) *Env {
	return &Env{
		homeDir:    homeDir,
		name:       name,
		binaryName: binaryName,
		Installer:  installer,
	}
}

type Installer interface {
	Install(version string, dstPath string) error
	Available() bool
}

func (env *Env) Use(version string) error {
	err := env.lock()
	if err != nil {
		return err
	}
	defer func() {
		_ = env.unlock()
	}()
	profilePath := env.profilePath()
	profile, err := env.profile()
	if err != nil {
		return err
	}
	if profile == nil {
		profile = &Profile{}
	}
	pv := &version
	if version == "" {
		pv = nil
	}
	profile.Version = pv
	profileContent, err := json.Marshal(profile)
	if err != nil {
		return err
	}
	if version == "" {
		return afero.WriteFile(Fs, profilePath, profileContent, os.ModePerm)
	}
	installed, err := env.Installed(version)
	if err != nil {
		return err
	}
	if !installed {
		err = env.Install(version)
		if err != nil {
			return err
		}
	}
	return afero.WriteFile(Fs, profilePath, profileContent, os.ModePerm)
}

func (env *Env) Installed(version string) (bool, error) {
	path := env.binaryPath(version)
	b, err := afero.Exists(Fs, path)
	if err != nil {
		return false, err
	}
	return b, nil
}

func (env *Env) Install(version string) error {
	installed, err := env.Installed(version)
	if err != nil {
		return err
	}
	if installed {
		return nil
	}
	return env.Installer.Install(version, env.binaryPath(version))
}

func (env *Env) ListInstalled() ([]string, error) {
	var installed []string
	dir, err := afero.ReadDir(Fs, filepath.Dir(env.profilePath()))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	for _, info := range dir {
		if info.IsDir() {
			installed = append(installed, info.Name())
		}
	}
	return installed, nil
}

func (env *Env) Name() string {
	return env.name
}

func (env *Env) BinaryName() string {
	return env.binaryName
}

func (env *Env) CurrentBinaryPath() (*string, error) {
	ver, err := env.CurrentVersion()
	if err != nil {
		return nil, err
	}
	if ver == nil {
		return nil, nil
	}

	binaryName := env.binaryName
	if Os == "windows" {
		binaryName = fmt.Sprintf("%s.exe", binaryName)
	}

	p := filepath.Join(env.homeDir, env.name, *ver, binaryName)
	return &p, nil
}

func (env *Env) CurrentVersion() (*string, error) {
	profile, err := env.profile()
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, nil
	}
	return profile.Version, nil
}

func (env *Env) Uninstall(version string) error {
	installed, err := env.Installed(version)
	if err != nil {
		return err
	}
	if !installed {
		return nil
	}
	err = Fs.RemoveAll(filepath.Dir(env.binaryPath(version)))
	if err != nil {
		return err
	}
	currentVersion, err := env.CurrentVersion()
	if err != nil {
		return err
	}
	if currentVersion != nil && *currentVersion == version {
		return env.Use("")
	}
	return nil
}

func (env *Env) binaryPath(version string) string {
	binaryName := env.binaryName
	if Os == "windows" {
		binaryName = fmt.Sprintf("%s.exe", binaryName)
	}
	return filepath.Join(env.homeDir, env.name, version, binaryName)
}

func (env *Env) profile() (*Profile, error) {
	exist, err := afero.Exists(Fs, env.profilePath())
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	profileContent, err := afero.ReadFile(Fs, env.profilePath())
	if err != nil {
		return nil, err
	}
	var profile Profile
	_ = json.Unmarshal(profileContent, &profile)
	return &profile, nil
}

func (env *Env) profilePath() string {
	return filepath.Join(env.homeDir, env.name, ".profile.json")
}

func (env *Env) lockPath() string {
	return filepath.Join(env.homeDir, env.name, ".lock")
}

func (env *Env) ensureHomeDir() {
	dir := filepath.Join(env.homeDir, env.name)
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		_ = os.MkdirAll(dir, 0755)
	}
}

func (env *Env) lock() error {
	if env.l != nil {
		return fmt.Errorf("this env has already been locked")
	}
	lockPath := env.lockPath()
	env.ensureHomeDir()
	lock := fslock.New(lockPath)
	err := lock.Lock()
	if err != nil {
		return err
	}
	env.l = lock
	return nil
}

func (env *Env) unlock() error {
	if env.l == nil {
		return nil
	}
	if err := env.l.Unlock(); err != nil {
		return err
	}
	env.l = nil
	return nil
}
