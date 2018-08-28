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
  alertTools:
    credentials:
      slack:
        apiurl: ""
        channel: ""
      victorOps:
        apikey: ""
        routingkey: ""
  domainName: kyma.local
  etcdBackupABS:
    containerName: ""
  isLocalEnv: true
  istio:
    tls:
      secretName: istio-ingress-certs
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  tlsCrt: ""
  tlsKey: ""
`

			installData, testOverrides := NewInstallationDataCreator().WithDomain("global.domainName", "kyma.local").WithLocalInstallation().GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when IP address is specified IsLocalEnv should be false", func() {

			const dummyOverridesForGlobal = `global:
  alertTools:
    credentials:
      slack:
        apiurl: ""
        channel: ""
      victorOps:
        apikey: ""
        routingkey: ""
  domainName: kyma.local
  etcdBackupABS:
    containerName: ""
  isLocalEnv: false
  istio:
    tls:
      secretName: istio-ingress-certs
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  tlsCrt: ""
  tlsKey: ""
`
			installData, testOverrides := NewInstallationDataCreator().WithDomain("global.domainName", "kyma.local").WithIP("100.100.100.100").GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when cert properties are provided tlsCrt and tlsKey should exist", func() {

			const dummyOverridesForGlobal = `global:
  alertTools:
    credentials:
      slack:
        apiurl: ""
        channel: ""
      victorOps:
        apikey: ""
        routingkey: ""
  domainName: kyma.local
  etcdBackupABS:
    containerName: ""
  isLocalEnv: false
  istio:
    tls:
      secretName: istio-ingress-certs
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  tlsCrt: abc
  tlsKey: def
`
			installData, testOverrides := NewInstallationDataCreator().WithDomain("global.domainName", "kyma.local").WithIP("100.100.100.100").WithCert("abc", "def").GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when remote env CA property is provided remoteEnvCa should exist", func() {

			const dummyOverridesForGlobal = `global:
  alertTools:
    credentials:
      slack:
        apiurl: ""
        channel: ""
      victorOps:
        apikey: ""
        routingkey: ""
  domainName: kyma.local
  etcdBackupABS:
    containerName: ""
  isLocalEnv: false
  istio:
    tls:
      secretName: istio-ingress-certs
  remoteEnvCa: xyz
  remoteEnvCaKey: abc
  tlsCrt: ""
  tlsKey: ""
`
			installData, testOverrides := NewInstallationDataCreator().WithDomain("global.domainName", "kyma.local").WithIP("100.100.100.100").WithRemoteEnvCa("xyz").WithRemoteEnvCaKey("abc").GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when EtcdBackupABSContainerName property is provided then etcdBackupABS.containerName should exist", func() {

			const dummyOverridesForGlobal = `global:
  alertTools:
    credentials:
      slack:
        apiurl: ""
        channel: ""
      victorOps:
        apikey: ""
        routingkey: ""
  domainName: kyma.local
  etcdBackupABS:
    containerName: abs/container/name
  isLocalEnv: false
  istio:
    tls:
      secretName: istio-ingress-certs
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  tlsCrt: ""
  tlsKey: ""
`
			installData, testOverrides := NewInstallationDataCreator().
				WithDomain("global.domainName", "kyma.local").
				WithIP("100.100.100.100").
				WithEtcdBackupABSContainerName("abs/container/name").
				GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)
			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})

		Convey("when slack and victorops credentials are provided then alertTools.credentials.victorOps and alertTools.credentials.slack should exist", func() {

			const dummyOverridesForGlobal = `global:
  alertTools:
    credentials:
      slack:
        apiurl: slack_apiurl
        channel: slack_channel
      victorOps:
        apikey: victorops_api_key
        routingkey: victorops_routing_key
  domainName: kyma.local
  etcdBackupABS:
    containerName: ""
  isLocalEnv: false
  istio:
    tls:
      secretName: istio-ingress-certs
  remoteEnvCa: ""
  remoteEnvCaKey: ""
  tlsCrt: ""
  tlsKey: ""
`
			installData, testOverrides := NewInstallationDataCreator().
				WithDomain("global.domainName", "kyma.local").
				WithIP("100.100.100.100").
				WithVictorOpsCredentials("victorops_routing_key", "victorops_api_key").
				WithSlackCredentials("slack_channel", "slack_apiurl").
				GetData()

			overridesMap, err := GetGlobalOverrides(&installData, UnflattenToMap(testOverrides))
			So(err, ShouldBeNil)

			overridesYaml, err := ToYaml(overridesMap)
			So(err, ShouldBeNil)
			So(overridesYaml, ShouldEqual, dummyOverridesForGlobal)
		})
	})
}
