package engine_test

import (
	"context"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	hookMock "github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/automock"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/engine"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestValidationEngine_Validate(t *testing.T) {
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

		validator := engine.NewTestValidator(webhook, time.Minute, fileReader)

		// When
		result, err := validator.Validate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(result.Success).To(gomega.Equal(true))
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

		validator := engine.NewTestValidator(webhook, time.Minute, fileReader)

		// When
		result, err := validator.Validate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(result.Success).To(gomega.Equal(true))
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

		validator := engine.NewTestValidator(webhook, time.Minute, fileReader)

		// When
		result, err := validator.Validate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(result.Success).To(gomega.Equal(true))
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

		validator := engine.NewTestValidator(webhook, time.Minute, fileReader)

		// When
		result, err := validator.Validate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(result.Success).To(gomega.Equal(false))
	})

	t.Run("CallError", func(t *testing.T) {
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
		webhook.On("Call", mock.Anything, "http://test.test-ns.svc.cluster.local/test", mock.Anything, mock.Anything).Return(errors.New("test"))
		defer webhook.AssertExpectations(t)

		validator := engine.NewTestValidator(webhook, time.Minute, fileReader)

		// When
		result, err := validator.Validate(ctx, accessor, "/tmp", files, services)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(result.Success).To(gomega.Equal(false))
	})
}
