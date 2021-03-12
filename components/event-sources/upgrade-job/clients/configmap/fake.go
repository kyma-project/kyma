package configmap

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func NewFakeClient(configMap *corev1.ConfigMap) (Client, error) {
	scheme, err := SetupSchemeOrDie()
	var dynamicClient *dynamicfake.FakeDynamicClient
	if err != nil {
		return Client{}, err
	}
	if configMap != nil {
		dynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, configMap)
	} else {
		dynamicClient = dynamicfake.NewSimpleDynamicClient(scheme)
	}
	return Client{DynamicClient: dynamicClient}, nil
}

func SetupSchemeOrDie() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return scheme, nil
}
