package controller

import (
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestFinalizers(t *testing.T) {
	// given
	p := protection{}
	var finalizers []string

	// then
	assert.False(t, p.hasFinalizer(finalizers))

	finalizers = p.addFinalizer(finalizers)
	assert.Contains(t, finalizers, v1alpha1.FinalizerAddonsConfiguration)
	assert.True(t, p.hasFinalizer(finalizers))

	finalizers = p.removeFinalizer(finalizers)
	assert.NotContains(t, finalizers, v1alpha1.FinalizerAddonsConfiguration)
	assert.False(t, p.hasFinalizer(finalizers))
}
