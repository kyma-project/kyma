package controller_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceSupervisorAggregatorGetSuccess(t *testing.T) {
	// given
	registeredKinds := []controller.Kind{
		controller.KindKubelessFunction,
		controller.KindDeployment,
	}

	// setup mock for all kinds
	mocks := make(map[controller.Kind]*automock.KubernetesResourceSupervisor)
	for _, k := range registeredKinds {
		m := &automock.KubernetesResourceSupervisor{}
		m.ExpectOnHasSynced(true).Once() // this method will be executed to ensure that given mock was called
		mocks[k] = m
	}

	// register all mock
	aggregator := controller.NewResourceSupervisorAggregator()
	for kind, mock := range mocks {
		aggregator.Register(kind, mock)
	}

	for _, k := range registeredKinds {
		// when
		concrete, err := aggregator.Get(k)

		// then
		require.NoError(t, err)
		assert.True(t, concrete.HasSynced())
		mocks[k].AssertExpectations(t)
	}
}

func TestResourceSupervisorAggregatorHasSynced(t *testing.T) {
	tests := map[string]struct {
		givenMocks map[controller.Kind]*automock.KubernetesResourceSupervisor
		expSynced  bool
	}{
		"all synced": {
			givenMocks: func() map[controller.Kind]*automock.KubernetesResourceSupervisor {
				registeredKinds := []controller.Kind{
					controller.KindKubelessFunction,
					controller.KindDeployment,
				}
				mocks := make(map[controller.Kind]*automock.KubernetesResourceSupervisor)
				for _, k := range registeredKinds {
					m := &automock.KubernetesResourceSupervisor{}
					m.ExpectOnHasSynced(true).Once()
					mocks[k] = m
				}
				return mocks
			}(),
			expSynced: true,
		},
		"one not synced": {
			givenMocks: func() map[controller.Kind]*automock.KubernetesResourceSupervisor {
				mocks := make(map[controller.Kind]*automock.KubernetesResourceSupervisor)
				d := &automock.KubernetesResourceSupervisor{}
				d.ExpectOnHasSynced(false).Once()
				mocks[controller.KindDeployment] = d

				fn := &automock.KubernetesResourceSupervisor{}
				fn.ExpectOnHasSynced(true).Once()
				mocks[controller.KindKubelessFunction] = fn
				return mocks
			}(),
			expSynced: false,
		},
		"all not synced": {
			givenMocks: func() map[controller.Kind]*automock.KubernetesResourceSupervisor {
				registeredKinds := []controller.Kind{
					controller.KindKubelessFunction,
					controller.KindDeployment,
				}
				mocks := make(map[controller.Kind]*automock.KubernetesResourceSupervisor)
				for _, k := range registeredKinds {
					m := &automock.KubernetesResourceSupervisor{}
					m.ExpectOnHasSynced(false).Once()
					mocks[k] = m
				}
				return mocks
			}(),
			expSynced: false,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			// register all mock
			aggregator := controller.NewResourceSupervisorAggregator()
			for kind, mock := range tc.givenMocks {
				aggregator.Register(kind, mock)
			}

			assert.Equal(t, tc.expSynced, aggregator.HasSynced())
		})
	}

}

func TestResourceSupervisorAggregatorGetNotFound(t *testing.T) {
	// given - empty aggregator
	aggregator := controller.NewResourceSupervisorAggregator()

	// when
	concrete, err := aggregator.Get(controller.KindDeployment)

	// then
	assert.True(t, controller.IsNotFoundError(err))
	assert.Nil(t, concrete)
}
