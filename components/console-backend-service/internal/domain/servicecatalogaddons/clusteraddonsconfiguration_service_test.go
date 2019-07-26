package servicecatalogaddons_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/helm-broker/pkg/apis/addons/v1alpha1"
	addonsFakeCli "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned/fake"
	addonsClientset "github.com/kyma-project/kyma/components/helm-broker/pkg/client/clientset/versioned/typed/addons/v1alpha1"
	addonsInformers "github.com/kyma-project/kyma/components/helm-broker/pkg/client/informers/externalversions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

func TestAddonsConfigurationService_AddRepos_Success(t *testing.T) {
	for tn, tc := range map[string]struct {
		name string
		urls []string
	}{
		"add URL": {
			name: "test",
			urls: []string{
				"www.next",
			},
		},
		"add many URLs": {
			name: "test",
			urls: []string{
				"www.next",
				"www.one",
				"www.two",
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			fixAddonCfg := fixClusterAddonsConfiguration("test")
			fixAddonCfg.Spec.Repositories = []v1alpha1.SpecRepository{
				{URL: "www.already.present.url"},
			}
			expURLs := append(tc.urls, "www.already.present.url")

			informer, client := fixClusterAddonsConfigurationInformer(fixAddonCfg)
			testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

			svc := servicecatalogaddons.NewAddonsConfigurationService(informer, client)

			// when
			result, err := svc.AddRepos(tc.name, tc.urls)

			// then
			require.NoError(t, err)
			assert.Len(t, result.Spec.Repositories, len(expURLs))

			var normalizedResultURLs []string
			for _, r := range result.Spec.Repositories {
				normalizedResultURLs = append(normalizedResultURLs, r.URL)
			}
			assert.ElementsMatch(t, expURLs, normalizedResultURLs)
		})
	}
}

func TestAddonsConfigurationService_AddRepos_Failure(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	fixURLs := []string{"www.www"}
	expErrMsg := fmt.Sprintf("%s doesn't exists", fixAddonCfgName)

	informer, client := fixClusterAddonsConfigurationInformer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	svc := servicecatalogaddons.NewAddonsConfigurationService(informer, client)

	// when
	result, err := svc.AddRepos(fixAddonCfgName, fixURLs)

	// then
	assert.EqualError(t, err, expErrMsg)
	assert.Nil(t, result)
}

func TestAddonsConfigurationService_DeleteRepos(t *testing.T) {
	for tn, tc := range map[string]struct {
		name         string
		repos        []v1alpha1.SpecRepository
		urlsToRemove []string
	}{
		"delete URL": {
			name: "test",
			repos: []v1alpha1.SpecRepository{
				{URL: "www.already.present.url"},
				{URL: "www.next"},
			},
			urlsToRemove: []string{
				"www.next",
			},
		},
		"delete many URLs": {
			name: "test",
			repos: []v1alpha1.SpecRepository{
				{URL: "www.already.present.url"},
				{URL: "www.next"},
				{URL: "www.second"},
			},
			urlsToRemove: []string{
				"www.next",
				"www.second",
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			cfg := fixClusterAddonsConfiguration(tc.name)
			cfg.Spec.Repositories = tc.repos

			inf, client := fixClusterAddonsConfigurationInformer(cfg)
			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client)

			// when
			result, err := svc.RemoveRepos(tc.name, tc.urlsToRemove)

			// then
			require.NoError(t, err)
			var normalizedResultURLs []string
			for _, r := range result.Spec.Repositories {
				normalizedResultURLs = append(normalizedResultURLs, r.URL)
			}
			for _, url := range tc.urlsToRemove {
				assert.NotContains(t, normalizedResultURLs, url)
			}
		})
	}
}

func TestAddonsConfigurationService_DeleteRepos_Failure(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	fixURLs := []string{"www.www"}
	expErrMsg := fmt.Sprintf("%s doesn't exists", fixAddonCfgName)

	inf, client := fixClusterAddonsConfigurationInformer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client)

	// when
	result, err := svc.RemoveRepos(fixAddonCfgName, fixURLs)

	// then
	assert.EqualError(t, err, expErrMsg)
	assert.Nil(t, result)
}

func TestAddonsConfigurationService_CreateAddonsConfiguration(t *testing.T) {
	for tn, tc := range map[string]struct {
		name   string
		urls   []string
		labels *gqlschema.Labels

		expectedResult *v1alpha1.ClusterAddonsConfiguration
	}{
		"successWithLabels": {
			name: "test",
			labels: &gqlschema.Labels{
				"add": "it",
				"ion": "al",
			},
			urls: []string{
				"ww.fix.k",
			},
			expectedResult: &v1alpha1.ClusterAddonsConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"add": "it",
						"ion": "al",
					},
				},
				Spec: v1alpha1.ClusterAddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{URL: "ww.fix.k"},
						},
					},
				},
			},
		},
		"successWithNilLabels": {
			name: "test",
			urls: []string{
				"ww.fix.k",
			},
			expectedResult: &v1alpha1.ClusterAddonsConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.ClusterAddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{URL: "ww.fix.k"},
						},
					},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			inf, client := fixClusterAddonsConfigurationInformer()
			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client)

			// when
			result, err := svc.Create(tc.name, tc.urls, tc.labels)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestAddonsConfigurationService_UpdateAddonsConfiguration(t *testing.T) {
	for tn, tc := range map[string]struct {
		name   string
		urls   []string
		labels *gqlschema.Labels

		expectedResult *v1alpha1.ClusterAddonsConfiguration
	}{
		"successWithLabels": {
			name: "test",
			labels: &gqlschema.Labels{
				"add": "it",
				"ion": "al",
			},
			urls: []string{
				"ww.fix.k",
			},
			expectedResult: &v1alpha1.ClusterAddonsConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"add": "it",
						"ion": "al",
					},
				},
				Spec: v1alpha1.ClusterAddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{URL: "ww.fix.k"},
						},
					},
				},
			},
		},
		"successWithNilLabels": {
			name: "test",
			urls: []string{
				"ww.fix.k",
			},
			expectedResult: &v1alpha1.ClusterAddonsConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.ClusterAddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{URL: "ww.fix.k"},
						},
					},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			inf, client := fixClusterAddonsConfigurationInformer(fixClusterAddonsConfiguration(tc.name))
			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client)

			// when
			result, err := svc.Update(tc.name, tc.urls, tc.labels)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestAddonsConfigurationService_DeleteAddonsConfiguration(t *testing.T) {
	// given
	fixAddonCfgName := "test"
	expectedCfg := fixClusterAddonsConfiguration(fixAddonCfgName)
	inf, client := fixClusterAddonsConfigurationInformer(expectedCfg)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client)

	// when
	cfg, err := svc.Delete(fixAddonCfgName)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func TestAddonsConfigurationService_DeleteAddonsConfiguration_Error(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	expErrMsg := fmt.Sprintf("%s doesn't exists", fixAddonCfgName)
	inf, client := fixClusterAddonsConfigurationInformer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client)

	// when
	cfg, err := svc.Delete(fixAddonCfgName)

	// then
	assert.EqualError(t, err, expErrMsg)
	assert.Nil(t, cfg)
}

func TestAddonsConfigurationService_ListAddonsConfigurations(t *testing.T) {
	for tn, tc := range map[string]struct {
		alreadyExitedCfgs  []runtime.Object
		expectedAddonsCfgs []*v1alpha1.ClusterAddonsConfiguration
	}{
		"empty": {
			alreadyExitedCfgs:  []runtime.Object{},
			expectedAddonsCfgs: []*v1alpha1.ClusterAddonsConfiguration(nil),
		},
		"few addons configurations": {
			alreadyExitedCfgs: []runtime.Object{
				fixClusterAddonsConfiguration("test"),
				fixClusterAddonsConfiguration("test2"),
			},
			expectedAddonsCfgs: []*v1alpha1.ClusterAddonsConfiguration{
				fixClusterAddonsConfiguration("test"),
				fixClusterAddonsConfiguration("test2"),
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			inf, _ := fixClusterAddonsConfigurationInformer(tc.alreadyExitedCfgs...)
			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil)

			// when
			result, err := svc.List(pager.PagingParams{})

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedAddonsCfgs, result)
		})
	}
}

func fixClusterAddonsConfiguration(name string) *v1alpha1.ClusterAddonsConfiguration {
	return &v1alpha1.ClusterAddonsConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.ClusterAddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{URL: "www.piko.bello"},
				},
			},
		},
	}
}

func fixClusterAddonsConfigurationInformer(objects ...runtime.Object) (cache.SharedIndexInformer, addonsClientset.AddonsV1alpha1Interface) {
	fakeCli := addonsFakeCli.NewSimpleClientset(objects...)
	addonsInformerFactory := addonsInformers.NewSharedInformerFactory(fakeCli, informerResyncPeriod)

	return addonsInformerFactory.Addons().V1alpha1().ClusterAddonsConfigurations().Informer(), fakeCli.AddonsV1alpha1()
}
