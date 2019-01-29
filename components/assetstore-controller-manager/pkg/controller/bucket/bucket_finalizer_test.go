package bucket_test

import (
	assetstorev1alpha1 "github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/bucket"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

import "testing"

func TestBucketFinalizerAdder_AddTo(t *testing.T) {
	//Given
	g := gomega.NewGomegaWithT(t)

	instance := &assetstorev1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Finalizers: []string{"test", "test2"},
		},
	}
	bucketFinalizer := bucket.NewBucketFinalizer()

	//When
	bucketFinalizer.AddTo(instance)

	//Then
	g.Expect(instance.Finalizers).To(gomega.ContainElement(bucket.DeleteBucketFinalizerName))
}

func TestBucketFinalizerAdder_DeleteFrom(t *testing.T) {
	//Given
	g := gomega.NewGomegaWithT(t)

	instance := &assetstorev1alpha1.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Finalizers: []string{"test", bucket.DeleteBucketFinalizerName, "test2"},
		},
	}
	bucketFinalizer := bucket.NewBucketFinalizer()

	//When
	bucketFinalizer.DeleteFrom(instance)

	//Then
	g.Expect(instance.Finalizers).NotTo(gomega.ContainElement(bucket.DeleteBucketFinalizerName))
}

func TestBucketFinalizerAdder_IsDefinedIn(t *testing.T) {
	t.Run("Defined", func(t *testing.T) {
		//Given
		g := gomega.NewGomegaWithT(t)

		instance := &assetstorev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{"foo", bucket.DeleteBucketFinalizerName, "bar"},
			},
		}
		bucketFinalizer := bucket.NewBucketFinalizer()

		//When
		result := bucketFinalizer.IsDefinedIn(instance)

		//Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("Defined", func(t *testing.T) {
		//Given
		g := gomega.NewGomegaWithT(t)

		instance := &assetstorev1alpha1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{"foo", "bar"},
			},
		}
		bucketFinalizer := bucket.NewBucketFinalizer()

		//When
		result := bucketFinalizer.IsDefinedIn(instance)

		//Then
		g.Expect(result).NotTo(gomega.BeTrue())
	})
}
