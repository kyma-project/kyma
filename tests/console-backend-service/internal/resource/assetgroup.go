package resource

import (
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type assetGroup struct {
	resCli *Resource
}

func NewAssetGroup(dynamicCli dynamic.Interface, namespace string, logFn func(format string, args ...interface{})) *assetGroup {
	return &assetGroup{
		resCli: New(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "assetgroups",
		}, namespace, logFn),
	}
}

func (self *assetGroup) Create(meta metav1.ObjectMeta, assetGroupSpec v1beta1.CommonAssetGroupSpec) error {
	assetGroup := &v1beta1.AssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: meta,
		Spec: v1beta1.AssetGroupSpec{
			CommonAssetGroupSpec: assetGroupSpec,
		},
	}

	err := self.resCli.Create(assetGroup)
	if err != nil {
		return errors.Wrapf(err, "while creating AssetGroup %s in namespace %s", meta.Name, meta.Namespace)
	}

	return err
}

func (self *assetGroup) Get(name string) (*v1beta1.AssetGroup, error) {
	u, err := self.resCli.Get(name)
	if err != nil {
		return nil, err
	}

	var res v1beta1.AssetGroup
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting AssetGroup %s", name)
	}

	return &res, nil
}

func (self *assetGroup) Delete(name string) error {
	err := self.resCli.Delete(name)
	if err != nil {
		return errors.Wrapf(err, "while deleting AssetGroup %s", name)
	}

	return nil
}
