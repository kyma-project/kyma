package rafter

import (
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/dynamicresource"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/waiter"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type clusterAssetGroup struct {
	resCli *dynamicresource.DynamicResource
	name   string
	spec   v1beta1.CommonAssetGroupSpec
}

func newClusterAssetGroup(dynamicCli dynamic.Interface, spec v1beta1.CommonAssetGroupSpec) *clusterAssetGroup {
	return &clusterAssetGroup{
		resCli: dynamicresource.NewClient(dynamicCli, schema.GroupVersionResource{
			Version:  v1beta1.GroupVersion.Version,
			Group:    v1beta1.GroupVersion.Group,
			Resource: "clusterassetgroups",
		}, ""),
		name: clusterAssetGroupName,
		spec: spec,
	}
}

func (ag *clusterAssetGroup) create() error {
	clusterAssetGroup := &v1beta1.ClusterAssetGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterAssetGroup",
			APIVersion: v1beta1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: ag.name,
		},
		Spec: v1beta1.ClusterAssetGroupSpec{
			CommonAssetGroupSpec: ag.spec,
		},
	}

	err := ag.resCli.Create(clusterAssetGroup)
	if err != nil {
		return errors.Wrapf(err, "while creating ClusterAssetGroup %s", ag.name)
	}

	return nil
}

func (ag *clusterAssetGroup) get() (*v1beta1.ClusterAssetGroup, error) {
	u, err := ag.resCli.Get(ag.name)
	if err != nil {
		return nil, err
	}

	var res v1beta1.ClusterAssetGroup
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &res)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ClusterAssetGroup %s", ag.name)
	}

	return &res, nil
}

func (ag *clusterAssetGroup) delete() error {
	err := ag.resCli.Delete(ag.name)
	if err != nil {
		return errors.Wrapf(err, "while deleting ClusterAssetGroup %s", ag.name)
	}

	return nil
}

func (ag *clusterAssetGroup) waitForStatusReady(stop <-chan struct{}) error {
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
		return errors.Wrapf(err, "while waiting for ready ClusterAssetGroup %s", ag.name)
	}

	return nil
}

func (ag *clusterAssetGroup) waitForRemove(stop <-chan struct{}) error {
	err := waiter.WaitAtMost(func() (bool, error) {
		_, err := ag.get()
		if err == nil {
			return false, nil
		}

		if !apierrors.IsNotFound(err) {
			return false, err
		}

		return true, nil
	}, waitTimeout, stop)
	if err != nil {
		return errors.Wrapf(err, "while waiting for delete ClusterAssetGroup %s", ag.name)
	}

	return err
}
