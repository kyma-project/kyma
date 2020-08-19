package servicecatalogaddons_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/helm-broker/pkg/apis/addons/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/servicecatalogaddons"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	testingUtils "github.com/kyma-project/kyma/components/console-backend-service/internal/testing"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/dynamic/dynamicinformer"
	bindingUsageApi "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

func TestAddonsConfigurationService_AddRepos_Success(t *testing.T) {
	for tn, tc := range map[string]struct {
		name string
		urls []*gqlschema.AddonsConfigurationRepositoryInput
	}{
		"add URL": {
			name: "test",
			urls: []*gqlschema.AddonsConfigurationRepositoryInput{
				{URL: "www.next"},
			},
		},
		"add many URLs": {
			name: "test",
			urls: []*gqlschema.AddonsConfigurationRepositoryInput{
				{URL: "www.next"},
				{URL: "www.one"},
				{URL: "www.two"},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			fixAddonCfg := fixAddonsConfiguration(tc.name)
			url := "www.already.present.url"
			fixAddonCfg.Spec.Repositories = []v1alpha1.SpecRepository{
				{URL: url},
			}
			expURLs := append(tc.urls, &gqlschema.AddonsConfigurationRepositoryInput{URL: url})

			client, err := newDynamicClient(fixAddonCfg)
			require.NoError(t, err)

			informer := fixAddonsConfigurationInformer(client)
			require.NoError(t, err)
			testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

			svc := servicecatalogaddons.NewAddonsConfigurationService(informer, client.Resource(addonsConfigGVR))

			// when
			result, err := svc.AddRepos(tc.name, "test", tc.urls)

			// then
			require.NoError(t, err)
			assert.Len(t, result.Spec.Repositories, len(expURLs))
		})
	}
}

func TestAddonsConfigurationService_AddRepos_Failure(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	fixURLs := []*gqlschema.AddonsConfigurationRepositoryInput{{URL: "www.www"}}

	client, err := newDynamicClient()
	require.NoError(t, err)

	informer := fixAddonsConfigurationInformer(client)
	require.NoError(t, err)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	svc := servicecatalogaddons.NewAddonsConfigurationService(informer, client.Resource(addonsConfigGVR))

	// when
	result, err := svc.AddRepos(fixAddonCfgName, "test", fixURLs)

	// then
	assert.Error(t, err)
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
			urlsToRemove: []string{"www.next"},
		},
		"delete many URLs": {
			name: "test",
			repos: []v1alpha1.SpecRepository{
				{URL: "www.already.present.url"},
				{URL: "www.next"},
				{URL: "www.second"},
			},
			urlsToRemove: []string{"www.next", "www.second"},
		},
		"delete all URLs": {
			name: "test",
			repos: []v1alpha1.SpecRepository{
				{URL: "www.already.present.url"},
				{URL: "www.next"},
				{URL: "www.second"},
			},

			urlsToRemove: []string{"www.already.present.url", "www.next", "www.second"},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			cfg := fixAddonsConfiguration(tc.name)
			cfg.Spec.Repositories = tc.repos

			client, err := newDynamicClient(cfg)
			require.NoError(t, err)

			inf := fixAddonsConfigurationInformer(client)
			require.NoError(t, err)
			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.Resource(addonsConfigGVR))

			// when
			result, err := svc.RemoveRepos(tc.name, "test", tc.urlsToRemove)

			// then
			require.NoError(t, err)
			assert.NotNil(t, result.Spec.Repositories)
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
	fixURLs := []string{"ww.fix.k"}
	client, err := newDynamicClient()
	require.NoError(t, err)

	inf := fixAddonsConfigurationInformer(client)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.Resource(addonsConfigGVR))

	// when
	result, err := svc.RemoveRepos(fixAddonCfgName, "test", fixURLs)

	// then
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAddonsConfigurationService_CreateAddonsConfiguration(t *testing.T) {
	for tn, tc := range map[string]struct {
		name   string
		urls   []*gqlschema.AddonsConfigurationRepositoryInput
		labels gqlschema.Labels

		expectedResult *v1alpha1.AddonsConfiguration
	}{
		"successWithLabels": {
			name: "test",
			labels: gqlschema.Labels{
				"add": "it",
				"ion": "al",
			},
			urls: []*gqlschema.AddonsConfigurationRepositoryInput{
				{
					URL: "ww.fix.k",
				},
			},
			expectedResult: &v1alpha1.AddonsConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AddonsConfiguration",
					APIVersion: "addons.kyma-project.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"add": "it",
						"ion": "al",
					},
				},
				Spec: v1alpha1.AddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{
								URL: "ww.fix.k",
							},
						},
					},
				},
			},
		},
		"successWithNilLabels": {
			name: "test",
			urls: []*gqlschema.AddonsConfigurationRepositoryInput{
				{
					URL: "ww.fix.k",
				},
			},
			expectedResult: &v1alpha1.AddonsConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "AddonsConfiguration",
					APIVersion: "addons.kyma-project.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.AddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{
								URL: "ww.fix.k",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			client, err := newDynamicClient()
			require.NoError(t, err)
			inf := fixAddonsConfigurationInformer(client)
			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.Resource(addonsConfigGVR))

			// when
			result, err := svc.Create(tc.name, "test", tc.urls, tc.labels)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestAddonsConfigurationService_UpdateAddonsConfiguration(t *testing.T) {
	for tn, tc := range map[string]struct {
		name   string
		urls   []*gqlschema.AddonsConfigurationRepositoryInput
		labels gqlschema.Labels

		expectedResult *v1alpha1.AddonsConfiguration
	}{
		"successWithLabels": {
			name: "test",
			labels: gqlschema.Labels{
				"add": "it",
				"ion": "al",
			},
			urls: []*gqlschema.AddonsConfigurationRepositoryInput{
				{
					URL: "ww.fix.k",
				},
			},
			expectedResult: &v1alpha1.AddonsConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "addonsconfiguration",
					APIVersion: "addons.kyma-project.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
					Labels: map[string]string{
						"add": "it",
						"ion": "al",
					},
				},
				Spec: v1alpha1.AddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{
								URL: "ww.fix.k",
							},
						},
					},
				},
			},
		},
		"successWithNilLabels": {
			name: "test",
			urls: []*gqlschema.AddonsConfigurationRepositoryInput{
				{
					URL: "ww.fix.k",
				},
			},
			expectedResult: &v1alpha1.AddonsConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "addonsconfiguration",
					APIVersion: "addons.kyma-project.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: v1alpha1.AddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{
								URL: "ww.fix.k",
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			client, err := newDynamicClient(fixAddonsConfiguration(tc.name))
			require.NoError(t, err)

			inf := fixAddonsConfigurationInformer(client)
			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.Resource(addonsConfigGVR))

			// when
			result, err := svc.Update(tc.name, "test", tc.urls, tc.labels)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestAddonsConfigurationService_DeleteAddonsConfiguration(t *testing.T) {
	// given
	fixAddonCfgName := "test"
	expectedCfg := fixAddonsConfiguration(fixAddonCfgName)
	client, err := newDynamicClient(expectedCfg)
	require.NoError(t, err)
	inf := fixAddonsConfigurationInformer(client)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.Resource(addonsConfigGVR))

	// when
	cfg, err := svc.Delete(fixAddonCfgName, "test")

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func TestAddonsConfigurationService_DeleteAddonsConfiguration_Error(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	expErrMsg := fmt.Sprintf("%s doesn't exists", fixAddonCfgName)
	client, err := newDynamicClient()
	require.NoError(t, err)
	inf := fixAddonsConfigurationInformer(client)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.Resource(addonsConfigGVR))

	// when
	cfg, err := svc.Delete(fixAddonCfgName, "test")

	// then
	assert.EqualError(t, err, expErrMsg)
	assert.Nil(t, cfg)
}

func TestAddonsConfigurationService_ListAddonsConfigurations(t *testing.T) {
	for tn, tc := range map[string]struct {
		alreadyExitedCfgs  []runtime.Object
		expectedAddonsCfgs []*v1alpha1.AddonsConfiguration
	}{
		"empty": {
			alreadyExitedCfgs:  []runtime.Object{},
			expectedAddonsCfgs: []*v1alpha1.AddonsConfiguration(nil),
		},
		"few addons configurations": {
			alreadyExitedCfgs: []runtime.Object{
				fixAddonsConfiguration("test"),
				fixAddonsConfiguration("test2"),
			},
			expectedAddonsCfgs: []*v1alpha1.AddonsConfiguration{
				fixAddonsConfiguration("test"),
				fixAddonsConfiguration("test2"),
			},
		},
		"wrong namespace": {
			alreadyExitedCfgs: []runtime.Object{
				&v1alpha1.AddonsConfiguration{
					TypeMeta: metav1.TypeMeta{
						Kind:       "addonsconfiguration",
						APIVersion: "addons.kyma-project.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "wrong",
					},
				},
			},
			expectedAddonsCfgs: nil,
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			client, err := newDynamicClient(tc.alreadyExitedCfgs...)
			require.NoError(t, err)

			inf := fixAddonsConfigurationInformer(client)
			require.NoError(t, err)

			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.Resource(clusterAddonsConfigGVR))

			// when
			result, err := svc.List("test", pager.PagingParams{})

			// then
			require.NoError(t, err)
			assert.ElementsMatch(t, tc.expectedAddonsCfgs, result)
		})
	}
}

func TestAddonsConfigurationService_ResyncAddonsConfiguration(t *testing.T) {
	// given
	fixAddonCfgName := "test"
	expectedCfg := fixAddonsConfiguration(fixAddonCfgName)

	client, err := newDynamicClient(expectedCfg)
	require.NoError(t, err)

	inf := fixAddonsConfigurationInformer(client)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.Resource(addonsConfigGVR))

	expectedCfg.Spec.ReprocessRequest = 1

	// when
	cfg, err := svc.Resync(fixAddonCfgName, expectedCfg.Namespace)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

var (
	usageKindsGVR = schema.GroupVersionResource{
		Version:  bindingUsageApi.SchemeGroupVersion.Version,
		Group:    bindingUsageApi.SchemeGroupVersion.Group,
		Resource: "usagekinds",
	}
	bindingUsageGVR = schema.GroupVersionResource{
		Version:  bindingUsageApi.SchemeGroupVersion.Version,
		Group:    bindingUsageApi.SchemeGroupVersion.Group,
		Resource: "servicebindingusages",
	}
	addonsConfigGVR = schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "addonsconfigurations",
	}
	clusterAddonsConfigGVR = schema.GroupVersionResource{
		Version:  v1alpha1.SchemeGroupVersion.Version,
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Resource: "clusteraddonsconfigurations",
	}
)

func fixAddonsConfiguration(name string) *v1alpha1.AddonsConfiguration {
	return &v1alpha1.AddonsConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "addonsconfiguration",
			APIVersion: "addons.kyma-project.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
		},
		Spec: v1alpha1.AddonsConfigurationSpec{
			CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
				Repositories: []v1alpha1.SpecRepository{
					{URL: "www.piko.bello"},
				},
			},
		},
	}
}

func fixAddonsConfigurationInformer(dynamic dynamic.Interface) cache.SharedIndexInformer {
	return dynamicinformer.NewDynamicSharedInformerFactory(dynamic, 10).ForResource(addonsConfigGVR).Informer()
}

func fixClusterAddonsConfigurationInformer(dynamic dynamic.Interface) cache.SharedIndexInformer {
	return dynamicinformer.NewDynamicSharedInformerFactory(dynamic, 10).ForResource(clusterAddonsConfigGVR).Informer()
}
