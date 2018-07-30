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
			overrides, err := GetCoreOverrides(&installationData)

			So(err, ShouldBeNil)
			So(overrides, ShouldBeBlank)
		})

		Convey("when InstallationData contains domain name overrides should contain yaml", func() {
			const dummyOverridesForCore = `
nginx-ingress:
  controller:
    service:
      loadBalancerIP: 1.1.1.1
configurations-generator:
  kubeConfig:
    clusterName: kyma.local
    url: 
    ca: 
cluster-users:
  users:
    adminGroup: testgroup
test:
  auth:
    username: ""
    password: ""
etcd-operator:
  backupOperator:
    enabled: ""
    abs:
      storageAccount: ""
      storageKey: ""
`
			installationData := NewInstallationDataCreator().WithDomain("kyma.local").WithRemoteEnvIP("1.1.1.1").WithAdminGroup("testgroup").
				GetData()

			overrides, err := GetCoreOverrides(&installationData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForCore)
		})

		Convey("when test properties are provided, auth.username and auth.password should exist", func() {
			const dummyOverridesForCore = `
nginx-ingress:
  controller:
    service:
      loadBalancerIP: 1.1.1.1
configurations-generator:
  kubeConfig:
    clusterName: kyma.local
    url: 
    ca: 
cluster-users:
  users:
    adminGroup: 
test:
  auth:
    username: "user1"
    password: "p@ssw0rd"
etcd-operator:
  backupOperator:
    enabled: ""
    abs:
      storageAccount: ""
      storageKey: ""
`
			installationData := NewInstallationDataCreator().
				WithDomain("kyma.local").
				WithRemoteEnvIP("1.1.1.1").
				WithUITestCredentials("user1", "p@ssw0rd").
				GetData()

			overrides, err := GetCoreOverrides(&installationData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForCore)
		})

		Convey("when etcd-operator properties are provided then enabled, abs.storageAccount and abs.storageKey should exist", func() {
			const dummyOverridesForCore = `
nginx-ingress:
  controller:
    service:
      loadBalancerIP: 1.1.1.1
configurations-generator:
  kubeConfig:
    clusterName: kyma.local
    url: 
    ca: 
cluster-users:
  users:
    adminGroup: 
test:
  auth:
    username: ""
    password: ""
etcd-operator:
  backupOperator:
    enabled: "true"
    abs:
      storageAccount: "pico-bello"
      storageKey: "123-456-3245-a23b"
`
			installationData := NewInstallationDataCreator().
				WithDomain("kyma.local").
				WithRemoteEnvIP("1.1.1.1").
				WithEtcdOperator("true", "pico-bello", "123-456-3245-a23b").
				GetData()

			overrides, err := GetCoreOverrides(&installationData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForCore)
		})
	})
}
