package application

import (
	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func NewFakeClient(applications *applicationconnectorv1alpha1.ApplicationList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	applicationsUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(applications)
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(applications)
	if err != nil {
		return Client{}, err
	}

	applicationsUnstructured := &unstructured.UnstructuredList{
		Object: applicationsUnstructuredMap,
		Items:  unstructuredItems,
	}
	applicationsUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, applicationsUnstructured)
	return Client{DynamicClient: dynamicClient}, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := applicationconnectorv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func toUnstructuredItems(applications *applicationconnectorv1alpha1.ApplicationList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, application := range applications.Items {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&application)
		if err != nil {
			return nil, err
		}
		applicationUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
		applicationUnstructured.SetGroupVersionKind(GroupVersionKind())
		unstructuredItems = append(unstructuredItems, *applicationUnstructured)
	}
	return unstructuredItems, nil
}
