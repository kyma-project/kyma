package rafter

import (
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/dynamicresource"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type assetGroup struct {
	resCli    *dynamicresource.DynamicResource
	name      string
	namespace string
	spec      v1beta1.CommonAssetGroupSpec
}

func newAssetGroup(dynamicCli dynamic.Interface, namespace string, spec v1beta1.CommonAssetGroupSpec) *assetGroup {
	return &assetGroup{
		resCli: dynamicresource.NewClient(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "assetgroups",
		}),
		namespace: namespace,
		name:      assetGroupName,
		spec:      spec,
	}
}

func (ag *assetGroup) create() error {
	assetGroup := &v1beta1.AssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ag.name,
			Namespace: ag.namespace,
		},
		Spec: v1beta1.AssetGroupSpec{
			CommonAssetGroupSpec: ag.spec,
		},
	}

	err := ag.resCli.Create(assetGroup)
	if err != nil {
		return errors.Wrapf(err, "while creating AssetGroup %s in namespace %s", ag.name, ag.namespace)
	}

	return nil
}

func (ag *assetGroup) get() (*v1beta1.AssetGroup, error) {
	var res v1beta1.AssetGroup
	err := ag.resCli.Get(ag.namespace, ag.name, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting AssetGroup %s in namespace %s", ag.name, ag.namespace)
	}

	return &res, nil
}

func (ag *assetGroup) waitForStatusReady(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		res, err := ag.get()
		if err != nil {
			return false, err
		}

		if res.Status.Phase != v1beta1.AssetGroupReady {
			return false, nil
		}

		return true, nil
	}, waitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for ready AssetGroup %s in namespace %s", ag.name, ag.namespace)
	}

	return nil
}
