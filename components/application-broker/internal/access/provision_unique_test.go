package access_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/application-broker/internal"
	"github.com/kyma-project/kyma/components/application-broker/internal/access"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage"
	"github.com/kyma-project/kyma/components/application-broker/internal/storage/driver/memory"
)

func TestUniquenessProvisionChecker(t *testing.T) {
	genStore := func() storage.Instance {
		iSeed := []struct{ iID, rsID, ns string }{
			{"instance-1", "service-A", "ns-prod"},
			{"instance-2", "service-A", "ns-stage"},
			{"instance-3", "service-C", "ns-prod"},
			{"instance-4", "service-C", "ns-prod"},
		}

		iStore := memory.NewInstance()
		for _, iS := range iSeed {
			iO := &internal.Instance{
				ID:        internal.InstanceID(iS.iID),
				ServiceID: internal.ServiceID(iS.rsID),
				Namespace: internal.Namespace(iS.ns),
			}
			require.NoError(t, iStore.Insert(iO))
		}

		return iStore
	}

	cases := []struct {
		iID, rsID, ns      string
		canProvisionOutput access.CanProvisionOutput
	}{
		{"instance-2", "service-A", "ns-stage", access.CanProvisionOutput{Allowed: true}},
		{"instance-4", "service-C", "ns-prod", access.CanProvisionOutput{Allowed: false, Reason: "already activated"}},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
			// GIVEN
			sut := access.NewUniquenessProvisionChecker(genStore())

			// WHEN
			gotCanProvisionOutput, err := sut.CanProvision(internal.InstanceID(tc.iID), internal.RemoteServiceID(tc.rsID), internal.Namespace(tc.ns))

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, gotCanProvisionOutput, tc.canProvisionOutput)
		})
	}
}
