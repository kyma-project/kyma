package engine_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	hookMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/automock"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine"
	engineMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine/automock"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func TestMutationEngine_Mutate(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		ctx := context.TODO()
		services := []v1alpha2.AssetWebhookService{
			{Name: "test", Namespace: "test-ns", Endpoint: "/test"},
		}
		files := []string{"test/a.txt", "test/b/c.txt"}

		accessor := mockAccessor("test", "test-ns", 1)
		defer accessor.AssertExpectations(t)

		webhook := new(hookMock.Webhook)
		webhook.On("Call", mock.Anything, "http://test.test-ns.svc.cluster.local/test", mock.Anything, mock.Anything).Return(nil)
		defer webhook.AssertExpectations(t)

		mutator := engine.NewTestMutator(webhook, time.Minute, fileReader, fileWriter)

		// When
		err := mutator.Mutate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
	})

	t.Run("NoServices", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		ctx := context.TODO()
		services := make([]v1alpha2.AssetWebhookService, 0)
		files := []string{"test/a.txt", "test/b/c.txt"}

		accessor := mockAccessor("test", "test-ns", 1)
		defer accessor.AssertExpectations(t)

		webhook := new(hookMock.Webhook)
		defer webhook.AssertExpectations(t)

		mutator := engine.NewTestMutator(webhook, time.Minute, fileReader, fileWriter)

		// When
		err := mutator.Mutate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
	})

	t.Run("NoFiles", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		ctx := context.TODO()
		services := []v1alpha2.AssetWebhookService{
			{Name: "test", Namespace: "test-ns", Endpoint: "/test"},
		}
		files := make([]string, 0)

		accessor := mockAccessor("test", "test-ns", 1)
		defer accessor.AssertExpectations(t)

		webhook := new(hookMock.Webhook)
		webhook.On("Call", mock.Anything, "http://test.test-ns.svc.cluster.local/test", mock.Anything, mock.Anything).Return(nil)
		defer webhook.AssertExpectations(t)

		mutator := engine.NewTestMutator(webhook, time.Minute, fileReader, fileWriter)

		// When
		err := mutator.Mutate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
	})

	t.Run("ReadError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		ctx := context.TODO()
		services := []v1alpha2.AssetWebhookService{
			{Name: "test", Namespace: "test-ns", Endpoint: "/test"},
		}
		files := []string{"test/a.txt", "test/b/rError.txt"}

		accessor := mockAccessor("test", "test-ns", 1)
		defer accessor.AssertExpectations(t)

		webhook := new(hookMock.Webhook)
		defer webhook.AssertExpectations(t)

		mutator := engine.NewTestMutator(webhook, time.Minute, fileReader, fileWriter)

		// When
		err := mutator.Mutate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("CallError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		ctx := context.TODO()
		services := []v1alpha2.AssetWebhookService{
			{Name: "test", Namespace: "test-ns", Endpoint: "/test"},
		}
		files := []string{"test/a.txt", "test/b/nope.txt"}

		accessor := mockAccessor("test", "test-ns", 1)
		defer accessor.AssertExpectations(t)

		webhook := new(hookMock.Webhook)
		webhook.On("Call", mock.Anything, "http://test.test-ns.svc.cluster.local/test", mock.Anything, mock.Anything).Return(errors.New("test"))
		defer webhook.AssertExpectations(t)

		mutator := engine.NewTestMutator(webhook, time.Minute, fileReader, fileWriter)

		// When
		err := mutator.Mutate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func mockAccessor(name, namespace string, times int) *engineMock.Accessor {
	accessor := new(engineMock.Accessor)
	accessor.On("GetNamespace").Return(namespace).Times(times)
	accessor.On("GetName").Return(name).Times(times)

	return accessor
}

func fileReader(filename string) ([]byte, error) {
	if strings.Contains(filename, "rError") {
		return nil, errors.New("test-error")
	}

	return []byte(filename), nil
}

func fileWriter(filename string, data []byte, perm os.FileMode) error {
	if strings.Contains(filename, "wError") {
		return errors.New("test-error")
	}

	return nil
}
