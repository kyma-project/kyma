package testenv

import (
	"fmt"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"log"
)

func StartTestWebservice(client dynamic.Interface, namespace, name string, port int) (string, error) {
	_, err := createConfigMap(client, namespace, name, port)
	if err != nil {
		return "", err
	}

	pod, err := createPod(client, namespace, name)
	if err != nil {
		deleteResource(client, "configmaps", namespace, name)
		return "", err
	}

	return fmt.Sprintf("%s:%d", pod.Status.HostIP, port), nil
}

func deleteResource(client dynamic.Interface, resource, namespace, name string) error {
	groupVersion := schema.GroupVersionResource{Group: "", Version: "v1", Resource: resource}
	return client.Resource(groupVersion).Namespace(namespace).Delete(name, nil)
}

func createConfigMap(client dynamic.Interface, namespace, name string, port int) (*v1.ConfigMap, error) {
	configMap := webserviceConfigMap(namespace, name, port)
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&configMap)
	if err != nil {
		return nil, errors.Wrap(err, "while converting ConfigMap to map[string]interface{}")
	}

	configMap = v1.ConfigMap{}
	err = create(client, resource, obj, &configMap)
	return &configMap, err
}

func createPod(client dynamic.Interface, namespace, name string) (*v1.Pod, error) {
	pod := webservicePod(namespace, name)
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&pod)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Pod to map[string]interface{}")
	}

	pod = v1.Pod{}
	err = create(client, resource, obj, &pod)
	return &pod, err
}

func create(client dynamic.Interface, resource schema.GroupVersionResource, unstructuredMap map[string]interface{}, obj interface{}) error {
	result, err := client.Resource(resource).Create(&unstructured.Unstructured{Object: unstructuredMap}, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "while creating resource")
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(result.Object, obj)
	if err != nil {
		return errors.Wrap(err, "while converting Unstructured resource")
	}

	return nil
}

func webservicePod(namespace, name string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{"sidecar.istio.io/inject": "false"},
			Labels:      map[string]string{"owner": "console-backend-service-tests"},
		},
		Spec: v1.PodSpec{
			Volumes: []v1.Volume{
				{
					Name: "config",
					VolumeSource: v1.VolumeSource{
						ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{
							Name: name,
						}},
					},
				},
			},
			Containers: []v1.Container{
				{
					Name:  "mockice",
					Image: "hudymi/mockice:0.1.1",
					Args:  []string{"--verbose", "--config", "/app/config.yaml"},
					VolumeMounts: []v1.VolumeMount{{
						Name:      "config",
						MountPath: "/app/config.yaml",
						ReadOnly:  true,
						SubPath:   "config.yaml",
					}},
				},
			},
		},
	}
}
func webserviceConfigMap(namespace, name string, port int) v1.ConfigMap {
	return v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{"sidecar.istio.io/inject": "false"},
			Labels:      map[string]string{"owner": "console-backend-service-tests"},
		},
		Data: map[string]string{
			"config.yaml": fmt.Sprintf(`
address: :%d
endpoints:
- name: README.md
  defaultResponseCode: 200
  defaultResponseContent: "# Test markdown"  
  defaultResponseContentType: text/markdown; charset=utf-8
`, port),
		},
	}
}

func StopTestWebservice(client dynamic.Interface, namespace, name string) {
	err := deleteResource(client, "configmaps", namespace, name)
	if err != nil {
		log.Println(fmt.Sprintf("Cannot delete ConfigMap %s/%s, because: %v", namespace, name, err))
	}
	err = deleteResource(client, "pods", namespace, name)
	if err != nil {
		log.Println(fmt.Sprintf("Cannot delete Pod %s/%s, because: %v", namespace, name, err))
	}
}
