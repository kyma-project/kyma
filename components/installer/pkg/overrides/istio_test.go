package overrides

import (
	"testing"

	. "github.com/kyma-project/kyma/components/installer/pkg/toolkit"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetIstioOverrides(t *testing.T) {
	Convey("GetIstioOverrides", t, func() {
		Convey("when IP address is not specified overrides should be empty", func() {

			installationData := NewInstallationDataCreator().WithEmptyIP().GetData()
			overrides, err := GetIstioOverrides(&installationData)

			So(err, ShouldBeNil)
			So(overrides, ShouldBeEmpty)
		})

		Convey("when IP address is specified should contain yaml", func() {
			const dummyOverridesForIstio = `
ingressgateway:
  service:
    externalPublicIp: 100.100.100.100
`
			installationData := NewInstallationDataCreator().WithIP("100.100.100.100").GetData()
			overrides, err := GetIstioOverrides(&installationData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForIstio)
		})
	})
}
