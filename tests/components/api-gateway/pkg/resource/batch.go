package resource

import (
	"log"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

type Batch struct {
	ResourceManager *Manager
	Mapper          *restmapper.DeferredDiscoveryRESTMapper
}

func (b *Batch) CreateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		resourceSchema, ns, _ := b.GetResourceSchemaAndNamespace(res)
		err := b.ResourceManager.CreateResource(k8sClient, resourceSchema, ns, res)
		if err != nil {
			return nil, err
		}
		gotRes, err = b.ResourceManager.GetResource(k8sClient, resourceSchema, ns, res.GetName())
		if err != nil {
			return nil, err
		}
	}
	return gotRes, nil
}

func (b *Batch) UpdateResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gotRes := &unstructured.Unstructured{}
	for _, res := range resources {
		resourceSchema, ns, _ := b.GetResourceSchemaAndNamespace(res)
		err := b.ResourceManager.UpdateResource(k8sClient, resourceSchema, ns, res.GetName(), res)
		if err != nil {
			return nil, err
		}
		gotRes, err = b.ResourceManager.GetResource(k8sClient, resourceSchema, ns, res.GetName())
		if err != nil {
			return nil, err
		}
	}
	return gotRes, nil
}

func (b *Batch) DeleteResources(k8sClient dynamic.Interface, resources ...unstructured.Unstructured) error {
	for _, res := range resources {
		resourceSchema, ns, name := b.GetResourceSchemaAndNamespace(res)
		err := b.ResourceManager.DeleteResource(k8sClient, resourceSchema, ns, name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Batch) GetResourceSchemaAndNamespace(manifest unstructured.Unstructured) (schema.GroupVersionResource, string, string) {
	namespace := manifest.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}
	resourceName := manifest.GetName()

	if manifest.GroupVersionKind().Kind == "Namespace" {
		namespace = ""
	}

	mapping, err := b.Mapper.RESTMapping(manifest.GroupVersionKind().GroupKind(), manifest.GroupVersionKind().Version)
	if err != nil {
		log.Fatal(err)
	}

	return mapping.Resource, namespace, resourceName
}
