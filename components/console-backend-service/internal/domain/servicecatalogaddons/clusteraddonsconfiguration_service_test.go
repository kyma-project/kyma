package servicecatalogaddons_test

import (
	"fmt"
	"strings"
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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	v12 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

const (
	addonsCfgKey        = "URLs"
	addonsCfgLabelValue = "true"
	addonsCfgLabelKey   = "helm-broker-repo"

	systemNs = "kyma-system"
)

func TestAddonsConfigurationService_AddRepos(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
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

				svc := servicecatalogaddons.NewAddonsConfigurationService(nil, informer, nil, client)

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
	})
	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		for tn, tc := range map[string]struct {
			name   string
			urls   []string
			errMsg *string
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
			"not existing addons config": {
				name:   "wrong",
				urls:   []string{"www.www"},
				errMsg: ptr(fmt.Sprintf("%s/wrong doesn't exists", systemNs)),
			},
		} {
			t.Run(tn, func(t *testing.T) {
				inf, client := fixConfigMapInformer(fixAddonsConfigMap("test"))
				testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

				svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil, client.ConfigMaps(systemNs), nil)

				result, err := svc.AddReposToConfigMap(tc.name, tc.urls)
				if tc.errMsg == nil {
					require.NoError(t, err)
					for _, url := range tc.urls {
						assert.Contains(t, result.Data[addonsCfgKey], url)
					}
				} else {
					assert.EqualError(t, err, *tc.errMsg)
				}

			})
		}
	})
}

func TestAddonsConfigurationService_AddRepos_Failure(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	fixURLs := []string{"www.www"}
	expErrMsg := fmt.Sprintf("%s doesn't exists", fixAddonCfgName)

	informer, client := fixClusterAddonsConfigurationInformer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, informer)

	svc := servicecatalogaddons.NewAddonsConfigurationService(nil, informer, nil, client)

	// when
	result, err := svc.AddRepos(fixAddonCfgName, fixURLs)

	// then
	assert.EqualError(t, err, expErrMsg)
	assert.Nil(t, result)
}

func TestAddonsConfigurationService_DeleteRepos(t *testing.T) {
	testURLs := []string{
		"www.next",
		"www.one",
		"www.two",
	}

	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
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

				svc := servicecatalogaddons.NewAddonsConfigurationService(nil, inf, nil, client)

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
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		for tn, tc := range map[string]struct {
			name   string
			urls   []string
			errMsg *string
		}{
			"delete URL": {
				name: "test",
				urls: []string{
					"www.next",
				},
			},
			"delete many URLs": {
				name: "test",
				urls: testURLs,
			},
			"not existing addons config": {
				name:   "wrong",
				urls:   []string{"www.www"},
				errMsg: ptr(fmt.Sprintf("%s/wrong doesn't exists", systemNs)),
			},
		} {
			t.Run(tn, func(t *testing.T) {
				cfg := fixAddonsConfigMap("test")
				cfg.Data[addonsCfgKey] = strings.Join(testURLs, "\n")
				inf, client := fixConfigMapInformer(cfg)
				testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

				svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil, client.ConfigMaps(systemNs), nil)

				result, err := svc.RemoveReposFromConfigMap(tc.name, tc.urls)
				if tc.errMsg != nil {
					assert.EqualError(t, err, *tc.errMsg)
				} else {
					require.NoError(t, err)
					for _, url := range tc.urls {
						assert.NotContains(t, result.Data[addonsCfgKey], url)
					}
				}
			})
		}
	})
}

func TestAddonsConfigurationService_DeleteRepos_Failure(t *testing.T) {
	// given
	fixAddonCfgName := "not-existing-cfg"
	fixURLs := []string{"www.www"}
	expErrMsg := fmt.Sprintf("%s doesn't exists", fixAddonCfgName)

	inf, client := fixClusterAddonsConfigurationInformer()
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(nil, inf, nil, client)

	// when
	result, err := svc.RemoveRepos(fixAddonCfgName, fixURLs)

	// then
	assert.EqualError(t, err, expErrMsg)
	assert.Nil(t, result)
}

func TestAddonsConfigurationService_CreateAddonsConfiguration(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
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

				svc := servicecatalogaddons.NewAddonsConfigurationService(nil, inf, nil, client)

				// when
				result, err := svc.Create(tc.name, tc.urls, tc.labels)

				// then
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			})
		}
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		for tn, tc := range map[string]struct {
			name   string
			urls   []string
			labels *gqlschema.Labels

			expectedResult *v1.ConfigMap
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
				expectedResult: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: systemNs,
						Labels: map[string]string{
							addonsCfgLabelKey: addonsCfgLabelValue,
							"add":             "it",
							"ion":             "al",
						},
					},
					Data: map[string]string{
						addonsCfgKey: "ww.fix.k",
					},
				},
			},
			"successWithNilLabels": {
				name: "test",
				urls: []string{
					"ww.fix.k",
				},
				expectedResult: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: systemNs,
						Labels: map[string]string{
							addonsCfgLabelKey: addonsCfgLabelValue,
						},
					},
					Data: map[string]string{
						addonsCfgKey: "ww.fix.k",
					},
				},
			},
		} {
			t.Run(tn, func(t *testing.T) {
				inf, client := fixConfigMapInformer()
				testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

				svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil, client.ConfigMaps(systemNs), nil)

				result, err := svc.CreateConfigMap(tc.name, tc.urls, tc.labels)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			})
		}
	})

}

func TestAddonsConfigurationService_UpdateAddonsConfiguration(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
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

				svc := servicecatalogaddons.NewAddonsConfigurationService(nil, inf, nil, client)

				// when
				result, err := svc.Update(tc.name, tc.urls, tc.labels)

				// then
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			})
		}
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		for tn, tc := range map[string]struct {
			name   string
			urls   []string
			labels *gqlschema.Labels

			expectedResult *v1.ConfigMap
			errMsg         *string
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
				expectedResult: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: systemNs,
						Labels: map[string]string{
							addonsCfgLabelKey: addonsCfgLabelValue,
							"add":             "it",
							"ion":             "al",
						},
					},
					Data: map[string]string{
						addonsCfgKey: "ww.fix.k",
					},
				},
			},
			"successWithNilLabels": {
				name: "test",
				urls: []string{
					"ww.fix.k",
				},
				expectedResult: &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: systemNs,
						Labels: map[string]string{
							addonsCfgLabelKey: addonsCfgLabelValue,
						},
					},
					Data: map[string]string{
						addonsCfgKey: "ww.fix.k",
					},
				},
			},
			"not existing addons config": {
				name:   "wrong",
				urls:   []string{"www.www"},
				errMsg: ptr(fmt.Sprintf("%s/wrong doesn't exists", systemNs)),
			},
		} {
			t.Run(tn, func(t *testing.T) {
				inf, client := fixConfigMapInformer(fixAddonsConfigMap("test"))
				testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

				svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil, client.ConfigMaps(systemNs), nil)

				result, err := svc.UpdateConfigMap(tc.name, tc.urls, tc.labels)

				if tc.errMsg != nil {
					assert.EqualError(t, err, *tc.errMsg)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tc.expectedResult, result)
				}
			})
		}
	})
}

func TestAddonsConfigurationService_DeleteAddonsConfiguration(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
		// given
		fixAddonCfgName := "test"
		expectedCfg := fixClusterAddonsConfiguration(fixAddonCfgName)
		inf, client := fixClusterAddonsConfigurationInformer(expectedCfg)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

		svc := servicecatalogaddons.NewAddonsConfigurationService(nil, inf, nil, client)

		// when
		cfg, err := svc.Delete(fixAddonCfgName)

		// then
		require.NoError(t, err)
		assert.Equal(t, expectedCfg, cfg)
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		expectedCfg := fixAddonsConfigMap("test")
		inf, client := fixConfigMapInformer(expectedCfg)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

		svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil, client.ConfigMaps(systemNs), nil)

		cfg, err := svc.DeleteConfigMap("test")
		require.NoError(t, err)
		assert.Equal(t, expectedCfg, cfg)
	})
}

func TestAddonsConfigurationService_DeleteAddonsConfiguration_Error(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
		// given
		fixAddonCfgName := "not-existing-cfg"
		expErrMsg := fmt.Sprintf("%s doesn't exists", fixAddonCfgName)
		inf, client := fixClusterAddonsConfigurationInformer()
		testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

		svc := servicecatalogaddons.NewAddonsConfigurationService(nil, inf, nil, client)

		// when
		cfg, err := svc.Delete(fixAddonCfgName)

		// then
		assert.EqualError(t, err, expErrMsg)
		assert.Nil(t, cfg)
	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		expectedCfg := fixAddonsConfigMap("test")
		inf, client := fixConfigMapInformer(expectedCfg)
		testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

		svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil, client.ConfigMaps(systemNs), nil)

		cfg, err := svc.DeleteConfigMap("wrong")
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}

func TestAddonsConfigurationService_ListAddonsConfigurations(t *testing.T) {
	t.Run("ClusterAddonsConfiguration", func(t *testing.T) {
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

				svc := servicecatalogaddons.NewAddonsConfigurationService(nil, inf, nil, nil)

				// when
				result, err := svc.List(pager.PagingParams{})

				// then
				require.NoError(t, err)
				assert.Equal(t, tc.expectedAddonsCfgs, result)
			})
		}

	})

	// testing deprecated scenario
	t.Run("ConfigMap", func(t *testing.T) {
		for tn, tc := range map[string]struct {
			givenCfgMaps       []runtime.Object
			expectedAddonsCfgs []*v1.ConfigMap
		}{
			"empty": {
				givenCfgMaps:       []runtime.Object{},
				expectedAddonsCfgs: []*v1.ConfigMap(nil),
			},
			"only addons repos": {
				givenCfgMaps: []runtime.Object{
					fixAddonsConfigMap("test"),
					fixAddonsConfigMap("test2"),
				},
				expectedAddonsCfgs: []*v1.ConfigMap{
					fixAddonsConfigMap("test"),
					fixAddonsConfigMap("test2"),
				},
			},
			"addons repos and excluded normal config maps": {
				givenCfgMaps: []runtime.Object{
					fixAddonsConfigMap("test"),
					fixConfigMap("excluded", "kyma-system"),
					fixConfigMap("excluded", "qa"),
				},
				expectedAddonsCfgs: []*v1.ConfigMap{
					fixAddonsConfigMap("test"),
				},
			},
		} {
			t.Run(tn, func(t *testing.T) {
				inf, _ := fixConfigMapInformer(tc.givenCfgMaps...)
				testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

				svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil, nil, nil)

				result, err := svc.ListConfigMaps(pager.PagingParams{})
				require.NoError(t, err)

				assert.Equal(t, tc.expectedAddonsCfgs, result)
			})
		}
	})
}

func fixConfigMap(name, namespace string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}

func fixAddonsConfigMap(name string) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: systemNs,
			Labels: map[string]string{
				"helm-broker-repo": "true",
			},
		},
		Data: map[string]string{
			addonsCfgKey: "www.piko.bo",
		},
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

func fixConfigMapInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := v12.NewFilteredConfigMapInformer(client, systemNs, time.Minute, cache.Indexers{}, func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", addonsCfgLabelKey, addonsCfgLabelValue)
	})

	return informerFactory, client.CoreV1()
}

func fixClusterAddonsConfigurationInformer(objects ...runtime.Object) (cache.SharedIndexInformer, addonsClientset.AddonsV1alpha1Interface) {
	fakeCli := addonsFakeCli.NewSimpleClientset(objects...)
	addonsInformerFactory := addonsInformers.NewSharedInformerFactory(fakeCli, informerResyncPeriod)

	return addonsInformerFactory.Addons().V1alpha1().ClusterAddonsConfigurations().Informer(), fakeCli.AddonsV1alpha1()
}

func ptr(s string) *string {
	return &s
}
