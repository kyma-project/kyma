package cms

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/dynamicresource"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type docsTopic struct {
	resCli    *dynamicresource.DynamicResource
	name      string
	namespace string
}

func newDocsTopic(dynamicCli dynamic.Interface, namespace string) *docsTopic {
	return &docsTopic{
		resCli: dynamicresource.NewClient(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha1.GroupVersion.Version,
			Group:    v1alpha1.GroupVersion.Group,
			Resource: "docstopics",
		}, namespace),
		namespace: namespace,
		name:      docsTopicName,
	}
}

func (dt *docsTopic) create(spec v1alpha1.CommonDocsTopicSpec) error {
	docsTopic := &v1alpha1.DocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DocsTopic",
			APIVersion: v1alpha1.GroupVersion.String(),
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
	}, waitTimeout, stop)
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
	}, waitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for delete DocsTopic %s in namespace %s", dt.name, dt.namespace)
	}

	return err
}
