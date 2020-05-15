package kymahelm_test

import (
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/kymahelm"
	"github.com/stretchr/testify/assert"
	helm "k8s.io/helm/pkg/proto/hapi/release"
	"testing"
)

func TestIsUpgradeStep(t *testing.T) {

	for _, tc := range []struct {
		status         kymahelm.ReleaseStatus
		expectedResult bool
		expectError    bool
	}{
		{
			status: kymahelm.ReleaseStatus{
				StatusCode:      helm.Status_PENDING_INSTALL,
				CurrentRevision: 1,
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			status: kymahelm.ReleaseStatus{
				StatusCode: helm.Status_PENDING_UPGRADE,
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			status: kymahelm.ReleaseStatus{
				StatusCode: helm.Status_DEPLOYED,
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			status: kymahelm.ReleaseStatus{
				StatusCode: helm.Status_PENDING_ROLLBACK,
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			status: kymahelm.ReleaseStatus{
				StatusCode:           helm.Status_FAILED,
				CurrentRevision:      2,
				LastDeployedRevision: 1,
			},
			expectedResult: true,
		},
		{
			status: kymahelm.ReleaseStatus{
				StatusCode:           helm.Status_UNKNOWN,
				CurrentRevision:      1,
				LastDeployedRevision: 0,
			},
			expectedResult: false,
		},
		{
			status: kymahelm.ReleaseStatus{
				StatusCode:           helm.Status_DELETED,
				CurrentRevision:      2,
				LastDeployedRevision: 0,
			},
			expectedResult: false,
			expectError:    true,
		},
		{
			status: kymahelm.ReleaseStatus{
				StatusCode: helm.Status_SUPERSEDED,
			},
			expectedResult: false,
			expectError:    true,
		},
	} {
		isUpgrade, err := tc.status.IsUpgradeStep()

		if tc.expectError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		assert.Equal(t, isUpgrade, tc.expectedResult)
	}
}
