package deployment

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func NewFakeClient(deployments *appsv1.DeploymentList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	deploymentsUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(deployments)
	if err != nil {
		return Client{}, err
	}

	unstructuredItems, err := toUnstructuredItems(deployments)
	if err != nil {
		return Client{}, err
	}

	deployUnstructured := &unstructured.UnstructuredList{
		Object: deploymentsUnstructuredMap,
		Items:  unstructuredItems,
	}
	deployUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, deployUnstructured)
	return Client{DynamicClient: dynamicClient}, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}

func toUnstructuredItems(deploymentList *appsv1.DeploymentList) ([]unstructured.Unstructured, error) {
	unstructuredItems := make([]unstructured.Unstructured, 0)
	for _, deployment := range deploymentList.Items {
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&deployment)
		if err != nil {
			return nil, err
		}
		deploymentUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
		deploymentUnstructured.SetGroupVersionKind(GroupVersionKind())
		unstructuredItems = append(unstructuredItems, *deploymentUnstructured)
	}
	return unstructuredItems, nil
}
