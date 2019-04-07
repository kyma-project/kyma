package resource

import (
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type docsTopic struct {
	resCli *Resource
}

func NewDocsTopic(dynamicCli dynamic.Interface, namespace string, logFn func(format string, args ...interface{})) *docsTopic {
	return &docsTopic{
		resCli: New(dynamicCli, schema.GroupVersionResource{
			Version:  v1alpha1.SchemeGroupVersion.Version,
			Group:    v1alpha1.SchemeGroupVersion.Group,
			Resource: "docstopics",
		}, namespace, logFn),
	}
}

func (dt *docsTopic) Create(meta metav1.ObjectMeta, docsTopicSpec v1alpha1.CommonDocsTopicSpec) error {
	docsTopic := &v1alpha1.DocsTopic{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DocsTopic",
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: meta,
		Spec: v1alpha1.DocsTopicSpec{
			CommonDocsTopicSpec: docsTopicSpec,
		},
	}

	err := dt.resCli.Create(docsTopic)
	if err != nil {
		return errors.Wrapf(err, "while creating DocsTopic %s in namespace %s", meta.Name, meta.Namespace)
	}

	return err
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
