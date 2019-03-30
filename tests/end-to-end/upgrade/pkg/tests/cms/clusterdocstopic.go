package cms

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/pkg/errors"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"k8s.io/client-go/dynamic"
)

type clusterDocsTopic struct {
	resCli      *resource.Resource
	name        string
}

func newClusterDocsTopicClient(dynamicCli dynamic.Interface) *clusterDocsTopic {
	return &clusterDocsTopic{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Resource: "clusterdocstopics",
		}, ""),
		name: ClusterDocsTopicName,
	}
}

func (dt *clusterDocsTopic) create(spec v1alpha1.CommonDocsTopicSpec) error {
	clusterDocsTopic := &v1alpha1.ClusterDocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterDocsTopic",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      dt.name,
		},
		Spec: v1alpha1.ClusterDocsTopicSpec{
			CommonDocsTopicSpec: spec,
		},
	}

	err := dt.resCli.Create(clusterDocsTopic)
	if err != nil {
		return errors.Wrapf(err, "while creating ClusterDocsTopic %s", dt.name)
	}

	return nil
}

func (dt *clusterDocsTopic) get() (*v1alpha1.ClusterDocsTopic, error) {
	u, err := dt.resCli.Get(dt.name)
	if err != nil {
		return nil, err
	}

	var res v1alpha1.ClusterDocsTopic
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterDocsTopic %s", dt.name)
	}

	return &res, nil
}

func (dt *clusterDocsTopic) delete() error {
	err := dt.resCli.Delete(dt.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterDocsTopic %s", dt.name)
	}

	return nil
}

func (dt *clusterDocsTopic) waitForStatusReady(stop <-chan struct{}) error {
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
		return errors.Wrapf(err, "while waiting for ready ClusterDocsTopic %s", dt.name)
	}

	return nil
}

func (dt *clusterDocsTopic) waitForRemove(stop <-chan struct{}) error {
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
		return errors.Wrapf(err, "while waiting for delete ClusterDocsTopic %s", dt.name)
	}

	return err
}