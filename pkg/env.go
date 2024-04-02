package pkg

type Env interface {
	Name() string
	BinaryName() string
	Installed(version string) (bool, error)
	Install(version string) error
	Uninstall(version string) error
	Use(version string) error
	CurrentVersion() (*string, error)
	CurrentBinaryPath() (*string, error)
}

func Use(env Env, version string) error {
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
	return env.Use(version)
}

func Uninstall(env Env, version string) error {
	err := env.Uninstall(version)
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
