package pkg

type Env interface {
	Installed(version string) bool
	Install(version string) error
	Use(version string) error
	CurrentVersion() *string
}

func Use(env Env, version string) error {
	if !env.Installed(version) {
		err := env.Install(version)
		if err != nil {
			return err
		}
	}
	return env.Use(version)
}
