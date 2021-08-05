package secret

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func NewFakeClient(secrets *corev1.SecretList) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	if err != nil {
		return Client{}, err
	}

	//deploymentsUnstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secrets)
	//if err != nil {
	//	return Client{}, err
	//}
	//
	//unstructuredItems, err := toUnstructuredItems(secrets)
	//if err != nil {
	//	return Client{}, err
	//}
	//
	//deployUnstructured := &unstructured.UnstructuredList{
	//	Object: deploymentsUnstructuredMap,
	//	Items:  unstructuredItems,
	//}
	//deployUnstructured.SetGroupVersionKind(GroupVersionKindList())

	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, secrets)
	return Client{client: dynamicClient}, nil
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

//func toUnstructuredItems(secretsList *corev1.SecretList) ([]unstructured.Unstructured, error) {
//	unstructuredItems := make([]unstructured.Unstructured, 0)
//	for _, secret := range secretsList.Items {
//		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&secret)
//		if err != nil {
//			return nil, err
//		}
//		deploymentUnstructured := &unstructured.Unstructured{Object: unstructuredMap}
//		deploymentUnstructured.SetGroupVersionKind(GroupVersionKind())
//		unstructuredItems = append(unstructuredItems, *deploymentUnstructured)
//	}
//	return unstructuredItems, nil
//}
