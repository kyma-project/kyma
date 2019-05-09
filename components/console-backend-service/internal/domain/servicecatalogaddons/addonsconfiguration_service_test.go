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

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.ConfigMaps(systemNs))

			result, err := svc.AddRepos(tc.name, tc.urls)
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
}

func TestAddonsConfigurationService_DeleteRepos(t *testing.T) {
	testURLs := []string{
		"www.next",
		"www.one",
		"www.two",
	}

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

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.ConfigMaps(systemNs))

			result, err := svc.RemoveRepos(tc.name, tc.urls)
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
}

func TestAddonsConfigurationService_CreateAddonsConfiguration(t *testing.T) {
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

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.ConfigMaps(systemNs))

			result, err := svc.Create(tc.name, tc.urls, tc.labels)
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

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.ConfigMaps(systemNs))

			result, err := svc.Update(tc.name, tc.urls, tc.labels)

			if tc.errMsg != nil {
				assert.EqualError(t, err, *tc.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}
		})
	}
}

func TestAddonsConfigurationService_DeleteAddonsConfiguration(t *testing.T) {
	expectedCfg := fixAddonsConfigMap("test")
	inf, client := fixConfigMapInformer(expectedCfg)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.ConfigMaps(systemNs))

	cfg, err := svc.Delete("test")
	require.NoError(t, err)
	assert.Equal(t, expectedCfg, cfg)
}

func TestAddonsConfigurationService_DeleteAddonsConfiguration_Error(t *testing.T) {
	expectedCfg := fixAddonsConfigMap("test")
	inf, client := fixConfigMapInformer(expectedCfg)
	testingUtils.WaitForInformerStartAtMost(t, time.Second, inf)

	svc := servicecatalogaddons.NewAddonsConfigurationService(inf, client.ConfigMaps(systemNs))

	cfg, err := svc.Delete("wrong")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestAddonsConfigurationService_ListAddonsConfigurations(t *testing.T) {
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

			svc := servicecatalogaddons.NewAddonsConfigurationService(inf, nil)

			result, err := svc.List(pager.PagingParams{})
			require.NoError(t, err)

			assert.Equal(t, tc.expectedAddonsCfgs, result)
		})
	}
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

func fixConfigMapInformer(objects ...runtime.Object) (cache.SharedIndexInformer, corev1.CoreV1Interface) {
	client := fake.NewSimpleClientset(objects...)
	informerFactory := v12.NewFilteredConfigMapInformer(client, systemNs, time.Minute, cache.Indexers{}, func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", addonsCfgLabelKey, addonsCfgLabelValue)
	})

	return informerFactory, client.CoreV1()
}

func ptr(s string) *string {
	return &s
}
