package application

import (
	"encoding/json"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type Client struct {
	DynamicClient dynamic.Interface
}

func NewClient(dInf dynamic.Interface) Client {
	return Client{DynamicClient: dInf}
}

func (c Client) List() (*applicationconnectorv1alpha1.ApplicationList, error) {
	applicationsUnstructured, err := c.DynamicClient.Resource(GroupVersionResource()).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return toApplicationList(applicationsUnstructured)
}

func (c Client) Get(namespace, name string) (*applicationconnectorv1alpha1.Application, error) {
	unstructuredApplication, err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return toApplication(unstructuredApplication)
}

func (c Client) Delete(namespace, name string) error {
	err := c.DynamicClient.Resource(GroupVersionResource()).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func toApplication(unstructuredDeployment *unstructured.Unstructured) (*applicationconnectorv1alpha1.Application, error) {
	application := new(applicationconnectorv1alpha1.Application)
	err := k8sruntime.DefaultUnstructuredConverter.FromUnstructured(unstructuredDeployment.Object, application)
	if err != nil {
		return nil, err
	}
	return application, nil
}

func toApplicationList(unstructuredList *unstructured.UnstructuredList) (*applicationconnectorv1alpha1.ApplicationList, error) {
	applicationList := new(applicationconnectorv1alpha1.ApplicationList)
	applicationListBytes, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(applicationListBytes, applicationList)
	if err != nil {
		return nil, err
	}
	return applicationList, nil
}

func GroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Version:  applicationconnectorv1alpha1.SchemeGroupVersion.Version,
		Group:    applicationconnectorv1alpha1.SchemeGroupVersion.Group,
		Resource: "applications",
	}
}

func GroupVersionKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: applicationconnectorv1alpha1.SchemeGroupVersion.Version,
		Group:   applicationconnectorv1alpha1.SchemeGroupVersion.Group,
		Kind:    "Application",
	}
}

func GroupVersionKindList() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Version: applicationconnectorv1alpha1.SchemeGroupVersion.Version,
		Group:   applicationconnectorv1alpha1.SchemeGroupVersion.Group,
		Kind:    "ApplicationList",
	}
}
