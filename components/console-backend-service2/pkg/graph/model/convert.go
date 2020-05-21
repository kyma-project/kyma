package model

import (
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (a *Application) FromUnstructured(u *unstructured.Unstructured) error {
	a.Name = u.GetName()
	return nil
}

type ApplicationList []*Application
func (l *ApplicationList) Append() resource.Convertible {
	converted := &Application{}
	*l = append(*l, converted)
	return converted
}

func (ns *Namespace) FromUnstructured(u *unstructured.Unstructured) error {
	ns.Name = u.GetName()
	ns.Labels = u.GetLabels()
	return nil
}

type NamespaceList []*Namespace
func (l *NamespaceList) Append() resource.Convertible {
	converted := &Namespace{}
	*l = append(*l, converted)
	return converted
}

