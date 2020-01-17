package resource

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Batch struct {
	ResourceManager *Manager
}

func (b *Batch) CreateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) {
	for _, res := range resources {
		resourceSchema, ns, _ := GetResourceSchemaAndNamespace(res)
		b.ResourceManager.CreateResource(k8sClient, resourceSchema, ns, res)
	}
}

func (b *Batch) UpdateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) {
	for _, res := range resources {
		resourceSchema, ns, _ := GetResourceSchemaAndNamespace(res)
		b.ResourceManager.UpdateResource(k8sClient, resourceSchema, ns, res.GetName(), res)
	}
}

func (b *Batch) DeleteResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) {
	for _, res := range resources {
		resourceSchema, ns, name := GetResourceSchemaAndNamespace(res)
		b.ResourceManager.DeleteResource(k8sClient, resourceSchema, ns, name)
	}
}

func GetResourceSchemaAndNamespace(manifest unstructured.Unstructured) (schema.GroupVersionResource, string, string) {
	metadata := manifest.Object["metadata"].(map[string]interface{})
	apiVersion := strings.Split(fmt.Sprintf("%s", manifest.Object["apiVersion"]), "/")
	namespace := "default"
	if metadata["namespace"] != nil {
		namespace = fmt.Sprintf("%s", metadata["namespace"])
	}
	resourceName := fmt.Sprintf("%s", metadata["name"])
	resourceKind := fmt.Sprintf("%s", manifest.Object["kind"])
	if resourceKind == "Namespace" {
		namespace = ""
	}
	//TODO: Move this ^ somewhere else and make it clearer
	apiGroup, version := getGroupAndVersion(apiVersion)
	resourceSchema := schema.GroupVersionResource{Group: apiGroup, Version: version, Resource: pluralForm(resourceKind)}
	return resourceSchema, namespace, resourceName
}

func getGroupAndVersion(apiVersion []string) (apiGroup string, version string) {
	if len(apiVersion) > 1 {
		return apiVersion[0], apiVersion[1]
	}
	return "", apiVersion[0]
}

func pluralForm(name string) string {
	return fmt.Sprintf("%ss", strings.ToLower(name))
}
