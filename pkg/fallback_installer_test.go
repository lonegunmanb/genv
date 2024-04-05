package pkg_test

import (
	"fmt"
	"testing"

	"github.com/lonegunmanb/genv/pkg"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestFallbackInstaller(t *testing.T) {
	cases := []struct {
		desc                     string
		installer1Success        bool
		installer2Success        bool
		expectedInstaller1Called bool
		expectedInstaller2Called bool
		installSuccess           bool
	}{
		{
			desc:                     "installer1 success",
			installer1Success:        true,
			installer2Success:        false,
			expectedInstaller1Called: true,
			expectedInstaller2Called: false,
			installSuccess:           true,
		},
		{
			desc:                     "installer2 success",
			installer1Success:        false,
			installer2Success:        true,
			expectedInstaller1Called: true,
			expectedInstaller2Called: true,
			installSuccess:           true,
		},
		{
			desc:                     "both installers fail",
			installer1Success:        false,
			installer2Success:        false,
			expectedInstaller1Called: true,
			expectedInstaller2Called: true,
			installSuccess:           false,
		},
	}
	for _, c := range cases {
		cc := c
		t.Run(cc.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockInstaller1 := NewMockInstaller(ctrl)
			mockInstaller2 := NewMockInstaller(ctrl)
			if cc.expectedInstaller1Called {
				if cc.installer1Success {
					mockInstaller1.EXPECT().Install("v1.0.0", "/tmp").Times(1).Return(nil)
				} else {
					mockInstaller1.EXPECT().Install("v1.0.0", "/tmp").Times(1).Return(fmt.Errorf("error"))
					mockInstaller1.EXPECT().Install("1.0.0", "/tmp").Times(1).Return(fmt.Errorf("error"))
				}
			}
			if cc.expectedInstaller2Called {
				if cc.installer2Success {
					mockInstaller2.EXPECT().Install("v1.0.0", "/tmp").Times(1).Return(nil)
				} else {
					mockInstaller2.EXPECT().Install("v1.0.0", "/tmp").Times(1).Return(fmt.Errorf("error"))
					mockInstaller2.EXPECT().Install("1.0.0", "/tmp").Times(1).Return(fmt.Errorf("error"))
				}
			}
			fi := pkg.NewFallbackInstaller(mockInstaller1, mockInstaller2)
			err := fi.Install("v1.0.0", "/tmp")
			if cc.installSuccess {
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestFallbackInstallerWillTryToFixVersionWhenSemverAndInstallError(t *testing.T) {
	cases := []string{
		"v1.0.0",
		"1.0.0",
	}
	for _, c := range cases {
		v := c
		t.Run(v, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockInstaller1 := NewMockInstaller(ctrl)
			mockInstaller2 := NewMockInstaller(ctrl)
			for _, ev := range cases {
				mockInstaller1.EXPECT().Install(ev, "/tmp").Times(1).Return(fmt.Errorf("error"))
			}
			mockInstaller2.EXPECT().Install(v, "/tmp").Times(1).Return(nil)
			sut := pkg.NewFallbackInstaller(mockInstaller1, mockInstaller2)
			err := sut.Install(v, "/tmp")
			assert.NoError(t, err)
		})
	}
}

func TestFallbackInstallerWontTryToFixVersionWhenNotSemverandInstallError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstaller1 := NewMockInstaller(ctrl)
	mockInstaller2 := NewMockInstaller(ctrl)
	v := "10ab64a0cd83ee20d259b6c0ecdfd785733ea2ee"
	mockInstaller1.EXPECT().Install(v, "/tmp").Times(1).Return(fmt.Errorf("error"))
	mockInstaller2.EXPECT().Install(v, "/tmp").Times(1).Return(nil)
	sut := pkg.NewFallbackInstaller(mockInstaller1, mockInstaller2)
	err := sut.Install(v, "/tmp")
	assert.NoError(t, err)
}
