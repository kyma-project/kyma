package assethook_test

import (
	"bytes"
	"context"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/api/v1alpha1"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/assethook/automock"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestWebhook_Call(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		httpClient := new(automock.HttpClient)
		webhook := assethook.New(httpClient)
		service := v1alpha2.WebhookService{Name: "tester", Namespace: "test", Endpoint: "/dummy"}
		ctx := context.TODO()

		request := bytes.NewBuffer([]byte("test-call"))
		response := &v1alpha1.ValidationResponse{}
		responseString := "{\"status\":{\"name1\":{\"status\":\"Failure\",\"message\":\"much more details of the failure\"},\"name2\":{\"status\":\"Success\",\"message\":\"much more details\"}}}"

		httpClient.On("Do", mock.Anything).Return(httpResponse(200, responseString), ctx.Err()).Once()
		defer httpClient.AssertExpectations(t)

		// When
		err := webhook.Do(ctx, "text/plain", service, request, response, time.Hour)

		// Then
		g.Expect(err).NotTo(gomega.HaveOccurred())
		g.Expect(response.Status).To(gomega.HaveLen(2))

		g.Expect(response.Status).To(gomega.HaveKey("name1"))
		g.Expect(response.Status).To(gomega.HaveKey("name2"))

		g.Expect(response.Status["name1"].Message).To(gomega.BeIdenticalTo("much more details of the failure"))
		g.Expect(response.Status["name2"].Message).To(gomega.BeIdenticalTo("much more details"))

		g.Expect(response.Status["name1"].Status).To(gomega.BeIdenticalTo(v1alpha1.ValidationFailure))
		g.Expect(response.Status["name2"].Status).To(gomega.BeIdenticalTo(v1alpha1.ValidationSuccess))
	})

	t.Run("Timeout", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		httpClient := new(automock.HttpClient)
		webhook := assethook.New(httpClient)
		service := v1alpha2.WebhookService{Name: "tester", Namespace: "test", Endpoint: "/dummy"}
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-20*time.Hour))
		defer cancel()

		request := bytes.NewBuffer([]byte("test-call"))
		response := &v1alpha1.ValidationResponse{}

		httpClient.On("Do", mock.Anything).Return(httpResponse(408, "{}"), ctx.Err()).Once()
		defer httpClient.AssertExpectations(t)

		// When
		err := webhook.Do(ctx, "text/plain", service, request, response, time.Hour)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("NotFound", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		httpClient := new(automock.HttpClient)
		webhook := assethook.New(httpClient)
		service := v1alpha2.WebhookService{Name: "tester", Namespace: "test", Endpoint: "/dummy"}
		ctx := context.TODO()

		request := bytes.NewBuffer([]byte("test-call"))
		response := &v1alpha1.ValidationResponse{}

		httpClient.On("Do", mock.Anything).Return(httpResponse(404, "{}"), nil).Once()
		defer httpClient.AssertExpectations(t)

		// When
		err := webhook.Do(ctx, "text/plain", service, request, response, time.Hour)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("ResponseParsingError", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		httpClient := new(automock.HttpClient)
		webhook := assethook.New(httpClient)
		service := v1alpha2.WebhookService{Name: "tester", Namespace: "test", Endpoint: "/dummy"}
		ctx := context.TODO()

		request := bytes.NewBuffer([]byte("test-call"))
		response := &v1alpha1.ValidationResponse{}

		httpClient.On("Do", mock.Anything).Return(httpResponse(200, "ala"), nil).Once()
		defer httpClient.AssertExpectations(t)

		// When
		err := webhook.Do(ctx, "text/plain", service, request, response, time.Hour)

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})
}

func httpResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
	}
}
