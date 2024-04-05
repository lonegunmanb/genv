package pkg

var _ Installer = &fallbackInstaller{}

type fallbackInstaller struct {
	i1 Installer
	i2 Installer
}

func (f *fallbackInstaller) Install(version string, dstPath string) error {
	err := f.i1.Install(version, dstPath)
	if err != nil {
		return f.i2.Install(version, dstPath)
	}
	return nil
}

func (f *fallbackInstaller) Available() bool {
	return f.i1.Available() || f.i2.Available()
}

func NewFallbackInstaller(i1 Installer, i2 Installer) Installer {
	return &fallbackInstaller{
		i1: i1,
		i2: i2,
	}
}
