// Code generated by failery v1.0.0. DO NOT EDIT.

package disabled

import resource "github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
import v1alpha1 "github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"

// docsTopicSvc is an autogenerated failing mock type for the docsTopicSvc type
type docsTopicSvc struct {
	err error
}

// NewDocsTopicSvc creates a new docsTopicSvc type instance
func NewDocsTopicSvc(err error) *docsTopicSvc {
	return &docsTopicSvc{err: err}
}

// Find provides a failing mock function with given fields: namespace, name
func (_m *docsTopicSvc) Find(namespace string, name string) (*v1alpha1.DocsTopic, error) {
	var r0 *v1alpha1.DocsTopic
	var r1 error
	r1 = _m.err

	return r0, r1
}

// Subscribe provides a failing mock function with given fields: listener
func (_m *docsTopicSvc) Subscribe(listener resource.Listener) {
}

// Unsubscribe provides a failing mock function with given fields: listener
func (_m *docsTopicSvc) Unsubscribe(listener resource.Listener) {
}
