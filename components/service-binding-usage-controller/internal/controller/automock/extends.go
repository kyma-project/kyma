package automock

import (
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	svcatSettings "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/settings/v1alpha1"
	"github.com/stretchr/testify/mock"
)

func (_m *KubernetesResourceSupervisor) ExpectOnEnsureLabelsCreated(ns string, resourceName string, usageName string, labels map[string]string) *mock.Call {
	return _m.On("EnsureLabelsCreated", ns, resourceName, usageName, labels).Return(nil)
}

func (_m *KindsSupervisors) ExpectOnGet(k controller.Kind, supervisor controller.KubernetesResourceSupervisor) *mock.Call {
	return _m.On("Get", k).Return(supervisor, nil)
}

func (_m *PodPresetModifier) ExpectOnUpsertPodPreset(newPodPreset *svcatSettings.PodPreset) *mock.Call {
	return _m.On("UpsertPodPreset", newPodPreset).Return(nil)
}

func (_m *BindingLabelsFetcher) ExpectOnFetch(inBinding *v1beta1.ServiceBinding, outLabels map[string]string) *mock.Call {
	return _m.On("Fetch", inBinding).Return(outLabels, nil)
}

func (_m *BindingLabelsFetcher) ExpectErrorOnFetch(outError error) *mock.Call {
	return _m.On("Fetch", mock.Anything).Return(nil, outError)
}

func (_m *BindingUsageChecker) ExpectOnValidateIfBindingUsageShouldBeProcessed(sbuFromRetry bool, bUsage *v1alpha1.ServiceBindingUsage) *mock.Call {
	return _m.On("ValidateIfBindingUsageShouldBeProcessed", sbuFromRetry, bUsage).Return(nil)
}

func (_m *BindingUsageChecker) ExpectErrorOnValidateIfBindingUsageShouldBeProcessed(sbuFromRetry bool, bUsage *v1alpha1.ServiceBindingUsage, err error) *mock.Call {
	return _m.On("ValidateIfBindingUsageShouldBeProcessed", sbuFromRetry, bUsage).Return(err)
}

func (_m *AppliedSpecStorage) ExpectOnGet(namespace, name string, spec *controller.UsageSpec, found bool) *mock.Call {
	return _m.On("Get", namespace, name).Return(spec, found, nil)
}

func (_m *AppliedSpecStorage) ExpectOnUpsert(bUsage *v1alpha1.ServiceBindingUsage, applied bool) *mock.Call {
	return _m.On("Upsert", bUsage, applied).Return(nil)
}

func (_m *BusinessMetric) ExpectOnRecordError(key string) *mock.Call {
	return _m.On("RecordError", key).Return(nil)
}

func (_m *BusinessMetric) ExpectOnIncrementQueueLength(key string) *mock.Call {
	return _m.On("IncrementQueueLength", key).Return(nil)
}

func (_m *BusinessMetric) ExpectOnDecrementQueueLength(key string) *mock.Call {
	return _m.On("DecrementQueueLength", key).Return(nil)
}

func (_m *BusinessMetric) ExpectOnRecordLatency(key string) *mock.Call {
	return _m.On("RecordLatency", key, mock.Anything).Return(nil)
}
