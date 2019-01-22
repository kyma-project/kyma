package bucket

import (
	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
)

const DeleteBucketFinalizerName = "deletebucket.finalizers.assetstore.kyma-project.io"

type bucketFinalizer struct{}

func NewBucketFinalizer() *bucketFinalizer {
	return &bucketFinalizer{}
}

func (f *bucketFinalizer) AddTo(instance *assetstorev1alpha1.Bucket) {
	instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, DeleteBucketFinalizerName)
}

func (f *bucketFinalizer) DeleteFrom(instance *assetstorev1alpha1.Bucket) {
	instance.ObjectMeta.Finalizers = f.removeString(instance.ObjectMeta.Finalizers, DeleteBucketFinalizerName)
}

func (f *bucketFinalizer) IsDefinedIn(instance *assetstorev1alpha1.Bucket) bool {
	return f.containsString(instance.ObjectMeta.Finalizers, DeleteBucketFinalizerName)
}

func (f *bucketFinalizer) containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func (f *bucketFinalizer) removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
