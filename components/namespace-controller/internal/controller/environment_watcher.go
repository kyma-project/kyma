package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/typed/core/v1"
)

type enviromentsWatcher struct {
	ns v1.NamespaceInterface
}

func (m enviromentsWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	return m.ns.List(listOptions)
}

func (m enviromentsWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return m.ns.Watch(listOptions)
}
