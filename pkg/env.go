package pkg

type Env interface {
	Name() string
	BinaryName() string
	Installed(version string) (bool, error)
	Install(version string) error
	Use(version string) error
	CurrentVersion() *string
	CurrentBinaryPath() (string, error)
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
