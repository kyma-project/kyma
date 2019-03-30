package cms

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/pkg/errors"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"k8s.io/client-go/dynamic"
)

type docsTopic struct {
	resCli      *resource.Resource
	name        string
	namespace   string
}

func newDocsTopic(dynamicCli dynamic.Interface, namespace string) *docsTopic {
	return &docsTopic{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Resource: "docstopics",
		}, namespace),
		namespace: namespace,
		name: DocsTopicName,
	}
}

func (dt *docsTopic) create(spec v1alpha1.CommonDocsTopicSpec) error {
	docsTopic := &v1alpha1.DocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DocsTopic",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dt.name,
			Namespace: dt.namespace,
		},
		Spec: v1alpha1.DocsTopicSpec{
			CommonDocsTopicSpec: spec,
		},
	}

	err := dt.resCli.Create(docsTopic)
	if err != nil {
		return errors.Wrapf(err, "while creating DocsTopic %s in namespace %s", dt.name, dt.namespace)
	}

	return nil
}

func (dt *docsTopic) get() (*v1alpha1.DocsTopic, error) {
	u, err := dt.resCli.Get(dt.name)
	if err != nil {
		return nil, err
	}

	var res v1alpha1.DocsTopic
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting DocsTopic %s in namespace %s", dt.name, dt.namespace)
	}

	return &res, nil
}

func (dt *docsTopic) delete() error {
	err := dt.resCli.Delete(dt.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting DocsTopic %s in namespace %s", dt.name, dt.namespace)
	}

	return nil
}

func (dt *docsTopic) waitForStatusReady(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		res, err := dt.get()
		if err != nil {
			return false, err
		}

		if res.Status.Phase != v1alpha1.DocsTopicReady {
			return false, nil
		}

		return true, nil
	}, WaitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready DocsTopic %s in namespace %s", dt.name, dt.namespace)
	}

	return nil
}

func (dt *docsTopic) waitForRemove(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		_, err := dt.get()
		if err == nil {
			return false, nil
		}

		if !apierrors.IsNotFound(err) {
			return false, err
		}

		return true, nil
	}, WaitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for delete DocsTopic %s in namespace %s", dt.name, dt.namespace)
	}

	return err
}
