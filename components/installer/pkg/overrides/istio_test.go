package overrides

import (
	"testing"

	. "github.com/kyma-project/kyma/components/installer/pkg/toolkit"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetIstioOverrides(t *testing.T) {
	Convey("GetIstioOverrides", t, func() {
		Convey("when IP address is not specified overrides should be empty", func() {

			installationData, testOverrides := NewInstallationDataCreator().GetData()
			overrides, err := GetIstioOverrides(&installationData, UnflattenToMap(testOverrides))

			So(err, ShouldBeNil)
			So(overrides, ShouldBeEmpty)
		})

		Convey("when IP address is specified should contain yaml", func() {
			const dummyOverridesForIstio = `gateways:
  istio-ingressgateway:
    service:
      externalPublicIp: 100.100.100.100
`
			installationData, testOverrides := NewInstallationDataCreator().WithIP("100.100.100.100").GetData()
			overridesMap, err := GetIstioOverrides(&installationData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForIstio)
		})
	})
}
