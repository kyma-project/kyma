package finalizer_test

import (
	"testing"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/finalizer"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testFinalizerName = "test.finalizer"

func TestFinalizerAdder_AddTo(t *testing.T) {
	//Given
	g := gomega.NewGomegaWithT(t)

	instance := &metav1.ObjectMeta{
		Finalizers: []string{"test", "test2"},
	}
	bucketFinalizer := finalizer.New(testFinalizerName)

	//When
	bucketFinalizer.AddTo(instance)
	bucketFinalizer.AddTo(instance)

	//Then
	g.Expect(instance.Finalizers).To(gomega.HaveLen(3))
	g.Expect(instance.Finalizers).To(gomega.ContainElement(testFinalizerName))
}

func TestFinalizerAdder_DeleteFrom(t *testing.T) {
	//Given
	g := gomega.NewGomegaWithT(t)

	instance := &metav1.ObjectMeta{
		Finalizers: []string{"test", testFinalizerName, "test2"},
	}
	bucketFinalizer := finalizer.New(testFinalizerName)

	//When
	bucketFinalizer.DeleteFrom(instance)

	//Then
	g.Expect(instance.Finalizers).NotTo(gomega.ContainElement(testFinalizerName))
}

func TestFinalizerAdder_IsDefinedIn(t *testing.T) {
	t.Run("Defined", func(t *testing.T) {
		//Given
		g := gomega.NewGomegaWithT(t)

		instance := &metav1.ObjectMeta{
			Finalizers: []string{"foo", testFinalizerName, "bar"},
		}
		bucketFinalizer := finalizer.New(testFinalizerName)

		//When
		result := bucketFinalizer.IsDefinedIn(instance)

		//Then
		g.Expect(result).To(gomega.BeTrue())
	})

	t.Run("NotDefined", func(t *testing.T) {
		//Given
		g := gomega.NewGomegaWithT(t)

		instance := &metav1.ObjectMeta{
			Finalizers: []string{"foo", "bar"},
		}

		bucketFinalizer := finalizer.New(testFinalizerName)

		//When
		result := bucketFinalizer.IsDefinedIn(instance)

		//Then
		g.Expect(result).NotTo(gomega.BeTrue())
	})
}
