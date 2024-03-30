package pkg

var _ Env = DownloadableEnv{}

type DownloadableEnv struct {
	downloadUrlTemplate string
	homeDir             string
}

func (d DownloadableEnv) Installed(version string) bool {
	//TODO implement me
	panic("implement me")
}

func (d DownloadableEnv) Install(version string) error {
	//TODO implement me
	panic("implement me")
}

func (d DownloadableEnv) Use(version string) error {
	//TODO implement me
	panic("implement me")
}

func (d DownloadableEnv) CurrentVersion() *string {
	//TODO implement me
	panic("implement me")
}
