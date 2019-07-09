package repository

import (
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryController(t *testing.T) {
	// Given
	tr := NewAddonsRepository("http://example.com/index.yaml")

	// Then
	assert.Equal(t, v1alpha1.RepositoryStatusReady, tr.Repository.Status)
}

func TestRepositoryController_IsFailed(t *testing.T) {
	// Given
	tr := NewAddonsRepository("http://example.com/index.yaml")

	// When
	tr.Failed()

	// Then
	assert.True(t, tr.IsFailed())
}
