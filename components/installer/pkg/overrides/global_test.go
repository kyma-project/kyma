package overrides

import (
	"testing"

	. "github.com/kyma-project/kyma/components/installer/pkg/toolkit"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGetGlobalOverrides(t *testing.T) {
	Convey("GetGlobalOverrides", t, func() {
		Convey("when IP address is not specified IsLocalEnv should be true", func() {
			const dummyOverridesForGlobal = `
global:
  tlsCrt: ""
  tlsKey: ""
  isLocalEnv: true
  domainName: "kyma.local"
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  istio:
    tls:
      secretName: "istio-ingress-certs"
  etcdBackupABS:
    containerName: ""
  alertTools:
    credentials:
      victorOps:
        routingkey: ""
        apikey: ""
      slack:
        channel: ""
        apiurl: ""
`

			installData := NewInstallationDataCreator().WithDomain("kyma.local").WithLocalInstallation().GetData()

			overrides, err := GetGlobalOverrides(&installData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when IP address is specified IsLocalEnv should be false", func() {
			const dummyOverridesForGlobal = `
global:
  tlsCrt: ""
  tlsKey: ""
  isLocalEnv: false
  domainName: "kyma.local"
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  istio:
    tls:
      secretName: "istio-ingress-certs"
  etcdBackupABS:
    containerName: ""
  alertTools:
    credentials:
      victorOps:
        routingkey: ""
        apikey: ""
      slack:
        channel: ""
        apiurl: ""
`
			installData := NewInstallationDataCreator().WithDomain("kyma.local").WithIP("100.100.100.100").GetData()

			overrides, err := GetGlobalOverrides(&installData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when cert properties are provided tlsCrt and tlsKey should exist", func() {
			const dummyOverridesForGlobal = `
global:
  tlsCrt: "abc"
  tlsKey: "def"
  isLocalEnv: false
  domainName: "kyma.local"
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  istio:
    tls:
      secretName: "istio-ingress-certs"
  etcdBackupABS:
    containerName: ""
  alertTools:
    credentials:
      victorOps:
        routingkey: ""
        apikey: ""
      slack:
        channel: ""
        apiurl: ""
`
			installData := NewInstallationDataCreator().WithDomain("kyma.local").WithIP("100.100.100.100").WithCert("abc", "def").GetData()

			overrides, err := GetGlobalOverrides(&installData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when remote env CA property is provided remoteEnvCa should exist", func() {
			const dummyOverridesForGlobal = `
global:
  tlsCrt: ""
  tlsKey: ""
  isLocalEnv: false
  domainName: "kyma.local"
  remoteEnvCa: "xyz"
  remoteEnvCaKey: "abc"
  istio:
    tls:
      secretName: "istio-ingress-certs"
  etcdBackupABS:
    containerName: ""
  alertTools:
    credentials:
      victorOps:
        routingkey: ""
        apikey: ""
      slack:
        channel: ""
        apiurl: ""
`
			installData := NewInstallationDataCreator().WithDomain("kyma.local").WithIP("100.100.100.100").WithRemoteEnvCa("xyz").WithRemoteEnvCaKey("abc").GetData()

			overrides, err := GetGlobalOverrides(&installData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when EtcdBackupABSContainerName property is provided then etcdBackupABS.containerName should exist", func() {
			const dummyOverridesForGlobal = `
global:
  tlsCrt: ""
  tlsKey: ""
  isLocalEnv: false
  domainName: "kyma.local"
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  istio:
    tls:
      secretName: "istio-ingress-certs"
  etcdBackupABS:
    containerName: "abs/container/name"
  alertTools:
    credentials:
      victorOps:
        routingkey: ""
        apikey: ""
      slack:
        channel: ""
        apiurl: ""
`
			installData := NewInstallationDataCreator().
				WithDomain("kyma.local").
				WithIP("100.100.100.100").
				WithEtcdBackupABSContainerName("abs/container/name").
				GetData()

			overrides, err := GetGlobalOverrides(&installData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when slack and victorops credentials are provided then alertTools.credentials.victorOps and alertTools.credentials.slack should exist", func() {
			const dummyOverridesForGlobal = `
global:
  tlsCrt: ""
  tlsKey: ""
  isLocalEnv: false
  domainName: "kyma.local"
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  istio:
    tls:
      secretName: "istio-ingress-certs"
  etcdBackupABS:
    containerName: ""
  alertTools:
    credentials:
      victorOps:
        routingkey: "victorops_routing_key"
        apikey: "victorops_api_key"
      slack:
        channel: "slack_channel"
        apiurl: "slack_apiurl"
`
			installData := NewInstallationDataCreator().
				WithDomain("kyma.local").
				WithIP("100.100.100.100").
				WithVictorOpsCredentials("victorops_routing_key", "victorops_api_key").
				WithSlackCredentials("slack_channel", "slack_apiurl").
				GetData()

			overrides, err := GetGlobalOverrides(&installData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForGlobal)
		})
	})
}
