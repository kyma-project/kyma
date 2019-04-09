package resource

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ClusterDocsTopic struct {
	resCli *Resource
}

func NewClusterDocsTopic(dynamicCli dynamic.Interface, logFn func(format string, args ...interface{})) *ClusterDocsTopic {
	return &ClusterDocsTopic{
		resCli: New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Resource: "clusterdocstopics",
		}, "", logFn),
	}
}

func (dt *ClusterDocsTopic) Create(meta metav1.ObjectMeta, docsTopicSpec v1alpha1.CommonDocsTopicSpec) error {
	docsTopic := &v1alpha1.ClusterDocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterDocsTopic",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: meta,
		Spec: v1alpha1.ClusterDocsTopicSpec{
			CommonDocsTopicSpec: docsTopicSpec,
		},
	}

	err := dt.resCli.Create(docsTopic)
	if err != nil {
		return errors.Wrapf(err, "while creating ClusterDocsTopic %s", meta.Name)
	}

	return err
}

func (dt *ClusterDocsTopic) Get(name string) (*v1alpha1.ClusterDocsTopic, error) {
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
