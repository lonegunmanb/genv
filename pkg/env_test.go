//go:generate mockgen -destination env_mock_test.go -package pkg_test . Env
package pkg_test

import (
	"github.com/lonegunmanb/genv/pkg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestEnvUseShouldCheckInstalled(t *testing.T) {
	version := "v0.1.0"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := NewMockEnv(ctrl)
	env.EXPECT().Installed(gomock.Eq(version)).Times(1).Return(false, nil)
	env.EXPECT().Install(gomock.Eq(version)).Times(1).Return(nil)
	env.EXPECT().Use(gomock.Eq(version)).Times(1).Return(nil)

	err := pkg.Use(env, version)
	assert.NoError(t, err)
}

func TestEnvShouldUseEmptyVersionAfterUninstallCurrentVersion(t *testing.T) {
	version := "v0.1.0"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := NewMockEnv(ctrl)
	env.EXPECT().CurrentVersion().Times(1).Return(&version, nil)
	env.EXPECT().Uninstall(gomock.Eq(version)).Times(1).Return(nil)
	env.EXPECT().Use("").Times(1).Return(nil)

	err := pkg.Uninstall(env, version)
	assert.NoError(t, err)
}

func TestEnvShouldNotUseEmptyVersionAfterUninstallNonCurrentVersion(t *testing.T) {
	currentVersion := "v0.1.0"
	uninstalledVersion := "v0.1.1"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := NewMockEnv(ctrl)
	env.EXPECT().CurrentVersion().Times(1).Return(&currentVersion, nil)
	env.EXPECT().Uninstall(gomock.Eq(uninstalledVersion)).Times(1).Return(nil)

	err := pkg.Uninstall(env, uninstalledVersion)
	assert.NoError(t, err)
}
