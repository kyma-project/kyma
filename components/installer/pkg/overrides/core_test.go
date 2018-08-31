package overrides

import (
	"testing"

	. "github.com/kyma-project/kyma/components/installer/pkg/toolkit"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCoreOverrides(t *testing.T) {
	Convey("GetCoreOverrides", t, func() {
		Convey("when InstallationData does not contain domain name overrides should be empty", func() {

			installationData, _ := NewInstallationDataCreator().GetData()
			overridesMap, err := GetCoreOverrides(&installationData, Map{})
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldBeBlank)
		})

		Convey("when InstallationData contains domain name overrides should contain yaml", func() {
			const dummyOverridesForCore = `configurations-generator:
  kubeConfig:
    clusterName: kyma.local
etcd-operator:
  backupOperator:
    abs:
      storageKey: ""
global:
  domainName: kyma.local
`
			installationData, testOverrides := NewInstallationDataCreator().WithGeneric("global.domainName", "kyma.local").WithGeneric("configurations-generator.kubeConfig.clusterName", "kyma.local").GetData()

			overridesMap, err := GetCoreOverrides(&installationData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForCore)
		})

		Convey("when test properties are provided, auth.username and auth.password should exist", func() {
			const dummyOverridesForCore = `configurations-generator:
  kubeConfig:
    clusterName: kyma.local
etcd-operator:
  backupOperator:
    abs:
      storageKey: ""
global:
  domainName: kyma.local
`
			installationData, testOverrides := NewInstallationDataCreator().
				WithGeneric("global.domainName", "kyma.local").
				WithGeneric("configurations-generator.kubeConfig.clusterName", "kyma.local").
				GetData()

			overridesMap, err := GetCoreOverrides(&installationData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForCore)
		})

		Convey("when etcd-operator properties are provided then enabled, abs.storageAccount and abs.storageKey should exist", func() {
			const dummyOverridesForCore = `configurations-generator:
  kubeConfig:
    clusterName: kyma.local
etcd-operator:
  backupOperator:
    abs:
      storageKey: ""
global:
  domainName: kyma.local
`
			installationData, testOverrides := NewInstallationDataCreator().
				WithGeneric("global.domainName", "kyma.local").
				WithGeneric("configurations-generator.kubeConfig.clusterName", "kyma.local").
				GetData()

			overridesMap, err := GetCoreOverrides(&installationData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForCore)
		})
	})
}
