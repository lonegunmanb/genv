package pkg

type Env interface {
	Installed(version string) bool
	Install(version string) error
	Use(version string) error
	CurrentVersion() *string
}
