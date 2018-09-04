package overrides

import (
	"testing"

	. "github.com/kyma-project/kyma/components/installer/pkg/toolkit"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetGlobalOverrides(t *testing.T) {
	Convey("GetGlobalOverrides", t, func() {
		Convey("when IP address is not specified IsLocalEnv should be true", func() {

			const dummyOverridesForGlobal = `global:
  domainName: kyma.local
  isLocalEnv: true
`

			installData, testOverrides := NewInstallationDataCreator().WithGeneric("global.domainName", "kyma.local").WithLocalInstallation().GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when IP address is specified IsLocalEnv should be false", func() {

			const dummyOverridesForGlobal = `global:
  domainName: kyma.local
  isLocalEnv: false
`
			installData, testOverrides := NewInstallationDataCreator().WithGeneric("global.domainName", "kyma.local").GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when cert properties are provided tlsCrt and tlsKey should exist", func() {

			const dummyOverridesForGlobal = `global:
  domainName: kyma.local
  isLocalEnv: false
`
			installData, testOverrides := NewInstallationDataCreator().WithGeneric("global.domainName", "kyma.local").GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})
	})
}
