package resource

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	controllerruntime "sigs.k8s.io/controller-runtime"

	"github.com/kyma-project/kyma/components/function-controller/internal/resource/automock"
)

func TestResourceSvc_Create(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		scheme := runtime.NewScheme()
		g.Expect(clientgoscheme.AddToScheme(scheme)).To(gomega.BeNil())

		parent := &batchv1.Job{}
		object := &corev1.Pod{}

		client := new(automock.Client)
		client.On("Create", ctx, object).Return(nil).Once()
		defer client.AssertExpectations(t)

		resourceClient := New(client, scheme)

		// When
		err := resourceClient.Create(ctx, parent, object)

		// Then
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("AlreadyExists", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		scheme := runtime.NewScheme()
		g.Expect(clientgoscheme.AddToScheme(scheme)).To(gomega.BeNil())

		parent := &batchv1.Job{}
		object := &corev1.Pod{}

		client := new(automock.Client)
		client.On("Create", ctx, object).Return(errors.NewAlreadyExists(controllerruntime.GroupResource{}, "test")).Once()
		defer client.AssertExpectations(t)

		resourceClient := New(client, scheme)

		// When
		err := resourceClient.Create(ctx, parent, object)

		// Then
		g.Expect(errors.IsAlreadyExists(err)).To(gomega.BeTrue())
	})

	t.Run("SetReferenceError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		scheme := runtime.NewScheme()

		parent := &batchv1.Job{}
		object := &corev1.Pod{}

		resourceClient := New(nil, scheme)

		// When
		err := resourceClient.Create(ctx, parent, object)

		// Then
		g.Expect(runtime.IsNotRegisteredError(err)).To(gomega.BeTrue())
	})
}

func TestResourceSvc_ListByLabel(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		list := &batchv1.JobList{}
		labels := map[string]string{"test": "test"}

		client := new(automock.Client)
		client.On("List", ctx, list, mock.Anything).Return(nil).Once()
		defer client.AssertExpectations(t)

		resourceClient := New(client, nil)

		// When
		err := resourceClient.ListByLabel(ctx, "test", labels, list)

		// Then
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("NoLabels", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		list := &batchv1.JobList{}

		client := new(automock.Client)
		client.On("List", ctx, list, mock.Anything).Return(nil).Once()
		defer client.AssertExpectations(t)

		resourceClient := New(client, nil)

		// When
		err := resourceClient.ListByLabel(ctx, "test", nil, list)

		// Then
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		list := &batchv1.JobList{}
		labels := map[string]string{"test": "test"}

		client := new(automock.Client)
		client.On("List", ctx, list, mock.Anything).Return(errors.NewBadRequest("bad")).Once()
		defer client.AssertExpectations(t)

		resourceClient := New(client, nil)

		// When
		err := resourceClient.ListByLabel(ctx, "test", labels, list)

		// Then
		g.Expect(errors.IsBadRequest(err)).To(gomega.BeTrue())
	})
}

func TestResourceSvc_DeleteAllBySelector(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		resourceType := &batchv1.Job{}
		labels := map[string]string{"test": "test"}
		selector := apilabels.SelectorFromSet(labels)

		client := new(automock.Client)
		client.On("DeleteAllOf", ctx, resourceType, mock.Anything).Return(nil).Once()
		defer client.AssertExpectations(t)

		resourceClient := New(client, nil)

		// When
		err := resourceClient.DeleteAllBySelector(ctx, resourceType, "test", selector)

		// Then
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("NoLabels", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		resourceType := &batchv1.Job{}

		client := new(automock.Client)
		client.On("DeleteAllOf", ctx, resourceType, mock.Anything).Return(nil).Once()
		defer client.AssertExpectations(t)

		resourceClient := New(client, nil)

		// When
		err := resourceClient.DeleteAllBySelector(ctx, resourceType, "test", nil)

		// Then
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("Error", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)
		resourceType := &batchv1.Job{}
		labels := map[string]string{"test": "test"}
		selector := apilabels.SelectorFromSet(labels)

		client := new(automock.Client)
		client.On("DeleteAllOf", ctx, resourceType, mock.Anything).Return(errors.NewBadRequest("bad")).Once()
		defer client.AssertExpectations(t)

		resourceClient := New(client, nil)

		// When
		err := resourceClient.DeleteAllBySelector(ctx, resourceType, "test", selector)

		// Then
		g.Expect(errors.IsBadRequest(err)).To(gomega.BeTrue())
	})
}
