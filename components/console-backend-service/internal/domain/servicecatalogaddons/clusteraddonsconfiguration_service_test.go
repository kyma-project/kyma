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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestClusterAddonsConfigurationService_AddRepos_Success(t *testing.T) {
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
			fixAddonCfg := fixClusterAddonsConfiguration("test")
			url := "www.already.present.url"
			fixAddonCfg.Spec.Repositories = []v1alpha1.SpecRepository{
				{URL: url},
			}
			expURLs := append(tc.urls, &gqlschema.AddonsConfigurationRepositoryInput{URL: url})

			client, err := newDynamicClient(fixAddonCfg)
			require.NoError(t, err)

			informer := fixClusterAddonsConfigurationInformer(client)
			require.NoError(t, err)

			testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

			svc := servicecatalogaddons.NewClusterAddonsConfigurationService(informer, client.Resource(clusterAddonsConfigGVR))

			// when
			result, err := svc.AddRepos(tc.name, tc.urls)

			// then
			require.NoError(t, err)
			assert.Len(t, result.Spec.Repositories, len(expURLs))
		})
	}
}

func TestClusterAddonsConfigurationService_AddRepos_Failure(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	fixURLs := []*gqlschema.AddonsConfigurationRepositoryInput{{URL: "www.www"}}

	client, err := newDynamicClient()
	require.NoError(t, err)

	informer := fixClusterAddonsConfigurationInformer(client)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	svc := servicecatalogaddons.NewClusterAddonsConfigurationService(informer, client.Resource(clusterAddonsConfigGVR))

	// when
	result, err := svc.AddRepos(fixAddonCfgName, fixURLs)

	// then
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestClusterAddonsConfigurationService_DeleteRepos(t *testing.T) {
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
			cfg := fixClusterAddonsConfiguration(tc.name)
			cfg.Spec.Repositories = tc.repos

			client, err := newDynamicClient(cfg)
			require.NoError(t, err)

			inf := fixClusterAddonsConfigurationInformer(client)
			require.NoError(t, err)

			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewClusterAddonsConfigurationService(inf, client.Resource(clusterAddonsConfigGVR))

			// when
			result, err := svc.RemoveRepos(tc.name, tc.urlsToRemove)

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

func TestClusterAddonsConfigurationService_DeleteRepos_Failure(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	fixURLs := []string{"ww.fix.k"}

	client, err := newDynamicClient()
	require.NoError(t, err)

	inf := fixClusterAddonsConfigurationInformer(client)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewClusterAddonsConfigurationService(inf, client.Resource(clusterAddonsConfigGVR))

	// when
	result, err := svc.RemoveRepos(fixAddonCfgName, fixURLs)

	// then
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestClusterAddonsConfigurationService_CreateAddonsConfiguration(t *testing.T) {
	for tn, tc := range map[string]struct {
		name   string
		urls   []*gqlschema.AddonsConfigurationRepositoryInput
		labels gqlschema.Labels

		expectedResult *v1alpha1.ClusterAddonsConfiguration
	}{
		"successWithLabels": {
			name: "test",
			labels: gqlschema.Labels{
				"add": "it",
				"ion": "al",
			},
			urls: []*gqlschema.AddonsConfigurationRepositoryInput{
				{
					URL:       "ww.fix.k",
					SecretRef: &gqlschema.ResourceRefInput{},
				},
			},
			expectedResult: &v1alpha1.ClusterAddonsConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterAddonsConfiguration",
					APIVersion: "addons.kyma-project.io/v1alpha1",
				},
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
							{
								URL:       "ww.fix.k",
								SecretRef: &v1.SecretReference{},
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
					URL:       "ww.fix.k",
					SecretRef: &gqlschema.ResourceRefInput{},
				},
			},
			expectedResult: &v1alpha1.ClusterAddonsConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterAddonsConfiguration",
					APIVersion: "addons.kyma-project.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.ClusterAddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{
								URL:       "ww.fix.k",
								SecretRef: &v1.SecretReference{},
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

			inf := fixClusterAddonsConfigurationInformer(client)
			require.NoError(t, err)

			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewClusterAddonsConfigurationService(inf, client.Resource(clusterAddonsConfigGVR))

			// when
			result, err := svc.Create(tc.name, tc.urls, tc.labels)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestClusterAddonsConfigurationService_UpdateAddonsConfiguration(t *testing.T) {
	for tn, tc := range map[string]struct {
		name   string
		urls   []*gqlschema.AddonsConfigurationRepositoryInput
		labels gqlschema.Labels

		expectedResult *v1alpha1.ClusterAddonsConfiguration
	}{
		"successWithLabels": {
			name: "test",
			labels: gqlschema.Labels{
				"add": "it",
				"ion": "al",
			},
			urls: []*gqlschema.AddonsConfigurationRepositoryInput{
				{
					URL:       "ww.fix.k",
					SecretRef: &gqlschema.ResourceRefInput{},
				},
			},
			expectedResult: &v1alpha1.ClusterAddonsConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterAddonsConfiguration",
					APIVersion: "addons.kyma-project.io/v1alpha1",
				},
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
							{
								URL:       "ww.fix.k",
								SecretRef: &v1.SecretReference{},
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
					URL:       "ww.fix.k",
					SecretRef: &gqlschema.ResourceRefInput{},
				},
			},
			expectedResult: &v1alpha1.ClusterAddonsConfiguration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterAddonsConfiguration",
					APIVersion: "addons.kyma-project.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: v1alpha1.ClusterAddonsConfigurationSpec{
					CommonAddonsConfigurationSpec: v1alpha1.CommonAddonsConfigurationSpec{
						Repositories: []v1alpha1.SpecRepository{
							{
								URL:       "ww.fix.k",
								SecretRef: &v1.SecretReference{},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			// given
			client, err := newDynamicClient(fixClusterAddonsConfiguration(tc.name))
			require.NoError(t, err)

			inf := fixClusterAddonsConfigurationInformer(client)
			require.NoError(t, err)

			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewClusterAddonsConfigurationService(inf, client.Resource(clusterAddonsConfigGVR))

			// when
			result, err := svc.Update(tc.name, tc.urls, tc.labels)

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}

func TestClusterAddonsConfigurationService_DeleteAddonsConfiguration(t *testing.T) {
	// given
	fixAddonCfgName := "test"
	expectedCfg := fixClusterAddonsConfiguration(fixAddonCfgName)
	client, err := newDynamicClient(expectedCfg)
	require.NoError(t, err)

	inf := fixClusterAddonsConfigurationInformer(client)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewClusterAddonsConfigurationService(inf, client.Resource(clusterAddonsConfigGVR))

	// when
	cfg, err := svc.Delete(fixAddonCfgName)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func TestClusterAddonsConfigurationService_DeleteAddonsConfiguration_Error(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	expErrMsg := fmt.Sprintf("%s doesn't exists", fixAddonCfgName)

	client, err := newDynamicClient()
	require.NoError(t, err)

	inf := fixClusterAddonsConfigurationInformer(client)
	require.NoError(t, err)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewClusterAddonsConfigurationService(inf, client.Resource(clusterAddonsConfigGVR))

	// when
	cfg, err := svc.Delete(fixAddonCfgName)

	// then
	assert.EqualError(t, err, expErrMsg)
	assert.Nil(t, cfg)
}

func TestClusterAddonsConfigurationService_ListAddonsConfigurations(t *testing.T) {
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
			client, err := newDynamicClient(tc.alreadyExitedCfgs...)
			require.NoError(t, err)

			inf := fixClusterAddonsConfigurationInformer(client)
			require.NoError(t, err)

			testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

			svc := servicecatalogaddons.NewClusterAddonsConfigurationService(inf, nil)

			// when
			result, err := svc.List(pager.PagingParams{})

			// then
			require.NoError(t, err)
			assert.Equal(t, tc.expectedAddonsCfgs, result)
		})
	}
}

func TestClusterAddonsConfigurationService_ResyncAddonsConfiguration(t *testing.T) {
	// given
	fixAddonCfgName := "test"
	expectedCfg := fixClusterAddonsConfiguration(fixAddonCfgName)

	client, err := newDynamicClient(expectedCfg)
	require.NoError(t, err)

	inf := fixClusterAddonsConfigurationInformer(client)

	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewClusterAddonsConfigurationService(inf, client.Resource(clusterAddonsConfigGVR))

	expectedCfg.Spec.ReprocessRequest = 1

	// when
	cfg, err := svc.Resync(fixAddonCfgName)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func fixClusterAddonsConfiguration(name string) *v1alpha1.ClusterAddonsConfiguration {
	return &v1alpha1.ClusterAddonsConfiguration{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAddonsConfiguration",
			APIVersion: "addons.kyma-project.io/v1alpha1",
		},
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
