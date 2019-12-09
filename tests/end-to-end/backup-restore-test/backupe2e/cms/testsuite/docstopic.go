package testsuite

import (
	"time"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/resource"
	"github.com/kyma-project/kyma/tests/end-to-end/backup-restore-test/utils/waiter"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type docsTopic struct {
	resCli      *resource.Resource
	name        string
	namespace   string
	waitTimeout time.Duration
}

func newDocsTopic(dynamicCli dynamic.Interface, name, namespace string, waitTimeout time.Duration) *docsTopic {
	return &docsTopic{
		resCli: resource.New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha1.GroupVersion.Version,
			Group:    v1alpha1.GroupVersion.Group,
			Resource: "docstopics",
		}, namespace),
		waitTimeout: waitTimeout,
		name:        name,
		namespace:   namespace,
	}
}

func (dt *docsTopic) Create(docsTopicSpec v1alpha1.CommonDocsTopicSpec) error {
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
			CommonDocsTopicSpec: docsTopicSpec,
		},
	}

	err := dt.resCli.Create(docsTopic)
	if err != nil {
		return errors.Wrapf(err, "while creating DocsTopic %s in namespace %s", dt.name, dt.namespace)
	}

	return err
}

func (dt *docsTopic) WaitForStatusReady() error {
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
		return errors.Wrapf(err, "while waiting for ready DocsTopic resources")
	}

	return nil
}

func (dt *docsTopic) Get(name string) (*v1alpha1.DocsTopic, error) {
	u, err := dt.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1alpha1.DocsTopic
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting DocsTopic %s", name)
	}

	return &res, nil
}
