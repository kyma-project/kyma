package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/resource"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/waiter"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type clusterDocsTopic struct {
	resCli      *resource.Resource
	name        string
	waitTimeout time.Duration
}

func newClusterDocsTopic(dynamicCli dynamic.Interface, name string, waitTimeout time.Duration) *clusterDocsTopic {
	return &clusterDocsTopic{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha1.GroupVersion.Version,
			Group:    v1alpha1.GroupVersion.Group,
			Resource: "clusterdocstopics",
		}, ""),
		waitTimeout: waitTimeout,
		name:        name,
	}
}

func (dt *clusterDocsTopic) Create(docsTopicSpec v1alpha1.CommonDocsTopicSpec) error {
	docsTopic := &v1alpha1.ClusterDocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterDocsTopic",
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: dt.name,
		},
		Spec: v1alpha1.ClusterDocsTopicSpec{
			CommonDocsTopicSpec: docsTopicSpec,
		},
	}

	err := dt.resCli.Create(docsTopic)
	if err != nil {
		return errors.Wrapf(err, "while creating ClusterDocsTopic %s", dt.name)
	}

	return err
}

func (dt *clusterDocsTopic) WaitForStatusReady() error {
	err := waiter.WaitAtMost(func() (bool, error) {
		res, err := dt.Get(dt.name)
		if err != nil {
			return false, err
		}

		if res.Status.Phase != v1alpha1.DocsTopicReady {
			return false, nil
		}

		return true, nil
	}, dt.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready ClusterDocsTopic resources")
	}

	return nil
}

func (dt *clusterDocsTopic) Get(name string) (*v1alpha1.ClusterDocsTopic, error) {
	u, err := dt.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1alpha1.ClusterDocsTopic
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterDocsTopic %s", name)
	}

	return &res, nil
}

func (dt *clusterDocsTopic) Delete() error {
	err := dt.resCli.Delete(dt.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterDocsTopic %s", dt.name)
	}

	return nil
}

func (dt *clusterDocsTopic) WaitForDeleted() error {
	err := waiter.WaitAtMost(func() (bool, error) {
		_, err := dt.Get(dt.name)
		if err == nil {
			return false, nil
		}

		if !apierrors.IsNotFound(err) {
			return false, err
		}

		return true, nil
	}, dt.waitTimeout)
	if err != nil {
		return errors.Wrapf(err, "while waiting for delete ClusterDocsTopic %s", dt.name)
	}

	return err
}
