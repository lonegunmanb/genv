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
			version := "v1.0.0"
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockInstaller1 := NewMockInstaller(ctrl)
			mockInstaller2 := NewMockInstaller(ctrl)
			if cc.expectedInstaller1Called {
				if cc.installer1Success {
					mockInstaller1.EXPECT().Install(version, "/tmp").Times(1).Return(nil)
				} else {
					mockInstaller1.EXPECT().Install(version, "/tmp").Times(1).Return(fmt.Errorf("error"))
				}
			}
			if cc.expectedInstaller2Called {
				if cc.installer2Success {
					mockInstaller2.EXPECT().Install(version, "/tmp").Times(1).Return(nil)
				} else {
					mockInstaller2.EXPECT().Install(version, "/tmp").Times(1).Return(fmt.Errorf("error"))
				}
			}
			fi := pkg.NewFallbackInstaller(mockInstaller1, mockInstaller2)
			err := fi.Install(version, "/tmp")
			if cc.installSuccess {
				assert.NoError(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
