package kymahelm_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	. "github.com/smartystreets/goconvey/convey"
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
							Status: kymahelm.StatusPendingInstall,
						},
						expectedResult: installation,
					},
					{
						status: kymahelm.ReleaseStatus{
							Status: kymahelm.StatusDeployed,
						},
						expectedResult: upgrade,
					},
					{
						status: kymahelm.ReleaseStatus{
							Status: kymahelm.StatusPendingUpgrade,
						},
						expectedResult: upgrade,
					},
					{
						status: kymahelm.ReleaseStatus{
							Status: kymahelm.StatusPendingRollback,
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

				statusCodesImportantPastRevisions := []kymahelm.Status{
					kymahelm.StatusFailed,
					kymahelm.StatusUnknown,
					kymahelm.StatusUninstalled,
					kymahelm.StatusUninstalling,
				}

				testInput := []struct {
					currentRevision      int
					lastDeployedRevision int
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
							Status:               testStatus,
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
			statusCodesNoProcessing := []kymahelm.Status{
				kymahelm.StatusSuperseded,
			}

			for _, testStatus := range statusCodesNoProcessing {

				//given
				statusObj := kymahelm.ReleaseStatus{
					Status: testStatus,
				}

				//when
				_, err := statusObj.IsUpgradeStep()

				//then
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "unexpected status")
			}
		})
	})
}
