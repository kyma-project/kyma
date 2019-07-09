package repository

import (
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryCollection_AddRepository(t *testing.T) {
	// Given
	trc := NewRepositoryCollection()

	// When
	trc.AddRepository(&RepositoryController{})
	trc.AddRepository(&RepositoryController{})

	// Then
	assert.Len(t, trc.Repositories, 2)
}

func TestRepositoryCollection_completeAddons(t *testing.T) {
	// Given
	trc := NewRepositoryCollection()

	// When
	trc.AddRepository(
		&RepositoryController{
			Addons: []*AddonController{
				{ID: "84e70958-5ae1-49b7-a78c-25983d1b3d0e"},
				{ID: ""},
				{ID: "2285fb92-3eb1-4e93-bc47-eacd40344c90"},
			},
		})
	trc.AddRepository(
		&RepositoryController{
			Addons: []*AddonController{
				{ID: "e89b4535-1728-4577-a6f6-e67998733a0f"},
				{ID: "ceabec68-30cf-40fc-b2d9-0d4cd24aee45"},
				{ID: ""},
			},
		})

	// Then
	assert.Len(t, trc.completeAddons(), 4)
}

func TestRepositoryCollection_ReadyAddons(t *testing.T) {
	// Given
	trc := NewRepositoryCollection()

	// When
	trc.AddRepository(
		&RepositoryController{
			Addons: []*AddonController{
				{
					ID:    "84e70958-5ae1-49b7-a78c-25983d1b3d0e",
					Addon: v1alpha1.Addon{Status: v1alpha1.AddonStatusReady},
				},
				{
					ID:    "2285fb92-3eb1-4e93-bc47-eacd40344c90",
					Addon: v1alpha1.Addon{Status: v1alpha1.AddonStatusReady},
				},
				{
					ID:    "e89b4535-1728-4577-a6f6-e67998733a0f",
					Addon: v1alpha1.Addon{Status: v1alpha1.AddonStatusFailed},
				},
				{
					ID:    "ceabec68-30cf-40fc-b2d9-0d4cd24aee45",
					Addon: v1alpha1.Addon{Status: v1alpha1.AddonStatusReady},
				},
			},
		})

	// Then
	assert.Len(t, trc.ReadyAddons(), 3)
}

func TestRepositoryCollection_IsRepositoriesIdConflict(t *testing.T) {

}

func TestRepositoryCollection_ReviseBundleDuplicationInRepository(t *testing.T) {

}

func TestRepositoryCollection_ReviseBundleDuplicationInStorage(t *testing.T) {

}
