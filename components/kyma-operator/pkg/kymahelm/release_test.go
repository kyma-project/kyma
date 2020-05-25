package kymahelm_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	. "github.com/smartystreets/goconvey/convey"
	helm "k8s.io/helm/pkg/proto/hapi/release"
)

func TestIsUpgradeStep(t *testing.T) {

	const installation = false
	const upgrade = true
	Convey("IsUpgradeStep", t, func() {
		Convey("should correctly detect upgradeable releases", func() {
			Convey("in cases depending only on release status", func() {

				testInput := []struct {
					status         kymahelm.ReleaseStatus
					expectedResult bool
				}{
					{
						status: kymahelm.ReleaseStatus{
							StatusCode: helm.Status_PENDING_INSTALL,
						},
						expectedResult: installation,
					},
					{
						status: kymahelm.ReleaseStatus{
							StatusCode: helm.Status_DEPLOYED,
						},
						expectedResult: upgrade,
					},
					{
						status: kymahelm.ReleaseStatus{
							StatusCode: helm.Status_PENDING_UPGRADE,
						},
						expectedResult: upgrade,
					},
					{
						status: kymahelm.ReleaseStatus{
							StatusCode: helm.Status_PENDING_ROLLBACK,
						},
						expectedResult: upgrade,
					},
				}

				for _, testData := range testInput {

					//given
					statusObj := testData.status

					//when
					isUpgrade, err := statusObj.IsUpgradeStep()

					//then
					So(err, ShouldBeNil)
					So(isUpgrade, ShouldEqual, testData.expectedResult)
				}
			})

			Convey("in cases depending on release status and it's past revisions", func() {

				statusCodesImportantPastRevisions := []helm.Status_Code{
					helm.Status_FAILED,
					helm.Status_UNKNOWN,
					helm.Status_DELETED,
					helm.Status_DELETING,
				}

				testInput := []struct {
					currentRevision      int32
					lastDeployedRevision int32
					expectedResult       bool
					expectedErrorMsg     string
				}{
					{
						currentRevision:      0,
						lastDeployedRevision: 0,
						expectedResult:       installation,
					},
					{
						currentRevision:      1,
						lastDeployedRevision: 0,
						expectedResult:       installation,
					},
					{
						//TODO: Is this correct?
						currentRevision:      1,
						lastDeployedRevision: 1,
						expectedResult:       installation,
					},
					{
						currentRevision:      2,
						lastDeployedRevision: 1,
						expectedResult:       upgrade,
					},
					{
						currentRevision:      3,
						lastDeployedRevision: 2,
						expectedResult:       upgrade,
					},
					{
						currentRevision:      2,
						lastDeployedRevision: 0,
						expectedErrorMsg:     "no deployed revision to rollback to",
					},
				}

				for _, testStatus := range statusCodesImportantPastRevisions {
					for _, testData := range testInput {

						//given
						statusObj := kymahelm.ReleaseStatus{
							StatusCode:           testStatus,
							CurrentRevision:      testData.currentRevision,
							LastDeployedRevision: testData.lastDeployedRevision,
						}

						//when
						isUpgrade, err := statusObj.IsUpgradeStep()

						//then
						if err != nil {
							So(err.Error(), ShouldContainSubstring, testData.expectedErrorMsg)
						} else {
							So(isUpgrade, ShouldEqual, testData.expectedResult)
						}
					}
				}
			})
		})
		Convey("should report an error for unrecognized status", func() {
			statusCodesNoProcessing := []helm.Status_Code{
				helm.Status_SUPERSEDED,
			}

			for _, testStatus := range statusCodesNoProcessing {

				//given
				statusObj := kymahelm.ReleaseStatus{
					StatusCode: testStatus,
				}

				//when
				_, err := statusObj.IsUpgradeStep()

				//then
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "unexpected status code")
			}
		})
	})
}
