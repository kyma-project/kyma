package overrides

import (
	"testing"

	. "github.com/kyma-project/kyma/components/installer/pkg/toolkit"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEnableAzureBroker(t *testing.T) {
	Convey("EnableAzureBroker", t, func() {
		Convey("when InstallationData does not contain azure credentials overrides should be empty", func() {

			installatioData := NewInstallationDataCreator().WithEmptyAzureCredentials().GetData()
			overrides, err := EnableAzureBroker(&installatioData)

			So(err, ShouldBeNil)
			So(overrides, ShouldBeBlank)
		})

		Convey("when InstallationData contains azure credentials overrides should contain yaml", func() {
			dummyOverridesForBroker := `
azure-broker:
  enabled: true
  subscription_id: d5423a63-0ab6-4455-9efe-569c6e716625
  tenant_id: 7ffdff3c-daa6-420d-9cff-b04769031acf
  client_id: 37bb544f-8935-4a00-a934-3999577fb637
  client_secret: ZGM3ZDlkYTgtZWMxMS00NTg4LTk5OGItOGU5YWJlNWUzYmE4DQo=
`
			installatioData := NewInstallationDataCreator().WithDummyAzureCredentials().GetData()
			overrides, err := EnableAzureBroker(&installatioData)

			So(err, ShouldBeNil)
			So(overrides, ShouldEqual, dummyOverridesForBroker)
		})
	})
}
