package controller_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	kindFunction   = controller.Kind("function")
	kindDeployment = controller.Kind("deployment")
)

func TestResourceSupervisorAggregatorGetSuccess(t *testing.T) {
	// given
	fnSupervisor := &automock.KubernetesResourceSupervisor{}
	deploySupervisor := &automock.KubernetesResourceSupervisor{}

	aggregator := controller.NewResourceSupervisorAggregator()
	aggregator.Register(kindFunction, fnSupervisor)
	aggregator.Register(kindDeployment, deploySupervisor)

	// when
	gotFn, err := aggregator.Get(kindFunction)
	require.NoError(t, err)
	gotDeploy, err := aggregator.Get(kindDeployment)
	require.NoError(t, err)

	// then
	assert.True(t, fnSupervisor == gotFn)
	assert.True(t, deploySupervisor == gotDeploy)
}

func TestResourceSupervisorAggregatorGetNotFound(t *testing.T) {
	// given - empty aggregator
	aggregator := controller.NewResourceSupervisorAggregator()

	// when
	concrete, err := aggregator.Get(kindDeployment)

	// then
	assert.True(t, controller.IsNotFoundError(err))
	assert.Nil(t, concrete)
}
