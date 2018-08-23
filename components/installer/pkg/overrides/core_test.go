package overrides

import (
	"testing"

	. "github.com/kyma-project/kyma/components/installer/pkg/toolkit"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetCoreOverrides(t *testing.T) {
	Convey("GetCoreOverrides", t, func() {
		Convey("when InstallationData does not contain domain name overrides should be empty", func() {

			installationData := NewInstallationDataCreator().WithEmptyDomain().GetData()
			overridesMap, err := GetCoreOverrides(&installationData)
			So(err, ShouldBeNil)

			overrides, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overrides, ShouldBeBlank)
		})

		Convey("when InstallationData contains domain name overrides should contain yaml", func() {
			const dummyOverridesForCore = `cluster-users:
  users:
    adminGroup: testgroup
configurations-generator:
  kubeConfig:
    ca: null
    clusterName: kyma.local
    url: null
etcd-operator:
  backupOperator:
    abs:
      storageAccount: ""
      storageKey: ""
    enabled: ""
nginx-ingress:
  controller:
    service:
      loadBalancerIP: 1.1.1.1
test:
  auth:
    password: ""
    username: ""
`
			installationData := NewInstallationDataCreator().WithDomain("kyma.local").WithRemoteEnvIP("1.1.1.1").WithAdminGroup("testgroup").
				GetData()

			overridesMap, err := GetCoreOverrides(&installationData)
			So(err, ShouldBeNil)

			overrides, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForCore)
		})

		Convey("when test properties are provided, auth.username and auth.password should exist", func() {
			const dummyOverridesForCore = `cluster-users:
  users:
    adminGroup: null
configurations-generator:
  kubeConfig:
    ca: null
    clusterName: kyma.local
    url: null
etcd-operator:
  backupOperator:
    abs:
      storageAccount: ""
      storageKey: ""
    enabled: ""
nginx-ingress:
  controller:
    service:
      loadBalancerIP: 1.1.1.1
test:
  auth:
    password: p@ssw0rd
    username: user1
`
			installationData := NewInstallationDataCreator().
				WithDomain("kyma.local").
				WithRemoteEnvIP("1.1.1.1").
				WithUITestCredentials("user1", "p@ssw0rd").
				GetData()

			overridesMap, err := GetCoreOverrides(&installationData)
			So(err, ShouldBeNil)

			overrides, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForCore)
		})

		Convey("when etcd-operator properties are provided then enabled, abs.storageAccount and abs.storageKey should exist", func() {
			const dummyOverridesForCore = `cluster-users:
  users:
    adminGroup: null
configurations-generator:
  kubeConfig:
    ca: null
    clusterName: kyma.local
    url: null
etcd-operator:
  backupOperator:
    abs:
      storageAccount: pico-bello
      storageKey: 123-456-3245-a23b
    enabled: "true"
nginx-ingress:
  controller:
    service:
      loadBalancerIP: 1.1.1.1
test:
  auth:
    password: ""
    username: ""
`
			installationData := NewInstallationDataCreator().
				WithDomain("kyma.local").
				WithRemoteEnvIP("1.1.1.1").
				WithEtcdOperator("true", "pico-bello", "123-456-3245-a23b").
				GetData()

			overridesMap, err := GetCoreOverrides(&installationData)
			So(err, ShouldBeNil)

			overrides, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForCore)
		})
	})
}
