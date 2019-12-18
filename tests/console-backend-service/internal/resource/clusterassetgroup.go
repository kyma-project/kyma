package resource

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ClusterAssetGroup struct {
	resCli *Resource
}

func NewClusterAssetGroup(dynamicCli dynamic.Interface, logFn func(format string, args ...interface{})) *ClusterAssetGroup {
	return &ClusterAssetGroup{
		resCli: New(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "clusterassetgroups",
		}, "", logFn),
	}
}

func (self *ClusterAssetGroup) Create(meta metav1.ObjectMeta, clusterAssetGroupSpec v1beta1.CommonAssetGroupSpec) error {
	clusterAssetGroup := &v1beta1.ClusterAssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: meta,
		Spec: v1beta1.ClusterAssetGroupSpec{
			CommonAssetGroupSpec: clusterAssetGroupSpec,
		},
	}

	err := self.resCli.Create(clusterAssetGroup)
	if err != nil {
		return errors.Wrapf(err, "while creating ClusterAssetGroup %s", meta.Name)
	}

	return err
}

func (self *ClusterAssetGroup) Get(name string) (*v1beta1.ClusterAssetGroup, error) {
	u, err := self.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1beta1.ClusterAssetGroup
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterAssetGroup %s", name)
	}

	return &res, nil
}

func (self *ClusterAssetGroup) Delete(name string) error {
	err := self.resCli.Delete(name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterAssetGroup %s", name)
	}

	return nil
}
