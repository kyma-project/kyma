package addons

import (
	"testing"

	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestRepositoryCollection_IsRepositoriesFailed(t *testing.T) {
	// Given
	trc := NewRepositoryCollection()

	// When
	trc.AddRepository(
		&RepositoryController{
			Repository: v1alpha1.StatusRepository{Status: v1alpha1.RepositoryStatusReady},
		})
	trc.AddRepository(
		&RepositoryController{
			Repository: v1alpha1.StatusRepository{Status: v1alpha1.RepositoryStatusReady},
		})

	// Then
	assert.False(t, trc.IsRepositoriesFailed())

	// When
	trc.AddRepository(&RepositoryController{
		Addons: []*AddonController{
			{
				Addon: v1alpha1.Addon{
					Status: v1alpha1.AddonStatusFailed,
				},
			},
		},
	})

	// Then
	assert.True(t, trc.IsRepositoriesFailed())
}

func TestRepositoryCollection_ReviseBundleDuplicationInRepository(t *testing.T) {
	// Given
	trc := NewRepositoryCollection()

	// When
	trc.AddRepository(
		&RepositoryController{
			Addons: []*AddonController{
				{
					ID:  "84e70958-5ae1-49b7-a78c-25983d1b3d0e",
					URL: "http://example.com/index.yaml",
					Addon: v1alpha1.Addon{
						Name:    "test",
						Version: "0.1",
						Status:  v1alpha1.AddonStatusReady,
					},
				},
				{
					ID:  "2285fb92-3eb1-4e93-bc47-eacd40344c90",
					URL: "http://example.com/index.yaml",
					Addon: v1alpha1.Addon{
						Name:    "test",
						Version: "0.2",
						Status:  v1alpha1.AddonStatusReady,
					},
				},
			},
		})
	trc.AddRepository(
		&RepositoryController{
			Addons: []*AddonController{
				{
					ID:  "e89b4535-1728-4577-a6f6-e67998733a0f",
					URL: "http://example.com/index-duplication.yaml",
					Addon: v1alpha1.Addon{
						Name:    "test",
						Version: "0.3",
						Status:  v1alpha1.AddonStatusReady,
					},
				},
				{
					ID:  "2285fb92-3eb1-4e93-bc47-eacd40344c90",
					URL: "http://example.com/index-duplication.yaml",
					Addon: v1alpha1.Addon{
						Name:    "test",
						Version: "0.4",
						Status:  v1alpha1.AddonStatusReady,
					},
				},
			},
		})
	trc.ReviseBundleDuplicationInRepository()

	// Then
	assert.Equal(t, string(v1alpha1.AddonStatusReady), string(findAddon(trc, "test", "0.1").Addon.Status))
	assert.Equal(t, string(v1alpha1.AddonStatusReady), string(findAddon(trc, "test", "0.2").Addon.Status))
	assert.Equal(t, string(v1alpha1.AddonStatusReady), string(findAddon(trc, "test", "0.3").Addon.Status))
	assert.Equal(t, string(v1alpha1.AddonStatusFailed), string(findAddon(trc, "test", "0.4").Addon.Status))
	assert.Equal(t,
		string(v1alpha1.AddonConflictInSpecifiedRepositories),
		string(findAddon(trc, "test", "0.4").Addon.Reason))
	assert.Equal(t,
		"Specified repositories have addons with the same ID: [url: http://example.com/index.yaml, addons: test:0.2]",
		string(findAddon(trc, "test", "0.4").Addon.Message))
}

func TestRepositoryCollection_ReviseBundleDuplicationInStorage(t *testing.T) {
	// Given
	trc := NewRepositoryCollection()
	list := &v1alpha1.AddonsConfigurationList{
		Items: []v1alpha1.AddonsConfiguration{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "addon-testing",
				},
				Status: v1alpha1.AddonsConfigurationStatus{
					CommonAddonsConfigurationStatus: v1alpha1.CommonAddonsConfigurationStatus{
						Repositories: []v1alpha1.StatusRepository{
							{
								URL: "http://example.com/index.yaml",
								Addons: []v1alpha1.Addon{
									{
										Name:    "test",
										Version: "0.2",
										Status:  v1alpha1.AddonStatusReady,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// When
	trc.AddRepository(
		&RepositoryController{
			Addons: []*AddonController{
				{
					ID:  "84e70958-5ae1-49b7-a78c-25983d1b3d0e",
					URL: "http://example.com/index.yaml",
					Addon: v1alpha1.Addon{
						Name:    "test",
						Version: "0.1",
						Status:  v1alpha1.AddonStatusReady,
					},
				},
				{
					ID:  "2285fb92-3eb1-4e93-bc47-eacd40344c90",
					URL: "http://example.com/index.yaml",
					Addon: v1alpha1.Addon{
						Name:    "test",
						Version: "0.2",
						Status:  v1alpha1.AddonStatusReady,
					},
				},
			},
		})
	trc.ReviseBundleDuplicationInStorage(list)

	// Then
	assert.Equal(t, string(v1alpha1.AddonStatusReady), string(findAddon(trc, "test", "0.1").Addon.Status))
	assert.Equal(t, string(v1alpha1.AddonStatusFailed), string(findAddon(trc, "test", "0.2").Addon.Status))
	assert.Equal(t,
		string(v1alpha1.AddonConflictWithAlreadyRegisteredAddons),
		string(findAddon(trc, "test", "0.2").Addon.Reason))
	assert.Equal(t,
		"An addon with the same ID is already registered: [ConfigurationName: addon-testing, url: http://example.com/index.yaml, addons: test:0.2]",
		string(findAddon(trc, "test", "0.2").Addon.Message))
}

func findAddon(rc *RepositoryCollection, name, version string) *AddonController {
	for _, addon := range rc.completeAddons() {
		if addon.Addon.Name == name && addon.Addon.Version == version {
			return addon
		}
	}

	return &AddonController{}
}
