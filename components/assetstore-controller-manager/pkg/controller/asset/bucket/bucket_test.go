package bucket_test

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/bucket"
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/controller/asset/bucket/fake"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestBucketService_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		expected := &v1alpha1.Bucket{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-obj",
				Namespace: "test-ns",
			},
		}

		informer := fake.NewInformer(expected)
		srv := bucket.New(informer)

		// When
		obj, err := srv.Get("test-ns", "test-obj")

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(obj).To(gomega.Equal(expected))
	})

	t.Run("NotFound", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		informer := fake.NewInformer()
		srv := bucket.New(informer)

		// When
		obj, err := srv.Get("test-ns", "not-found")

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(obj).To(gomega.BeNil())
	})

	t.Run("InvalidType", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		expected := &v1alpha1.Asset{
			ObjectMeta: v1.ObjectMeta{
				Name:      "test-obj",
				Namespace: "test-ns",
			},
		}

		expected.GetObjectMeta()

		informer := fake.NewInformer(expected)
		srv := bucket.New(informer)

		// When
		obj, err := srv.Get("test-ns", "test-obj")

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(obj).To(gomega.BeNil())
	})

	t.Run("InvalidType", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		informer := fake.NewInformer()
		srv := bucket.New(informer)

		// When
		obj, err := srv.Get("", "test-obj")

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(obj).To(gomega.BeNil())
	})
}
