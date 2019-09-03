package assethook_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/assethook"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/internal/assethook/automock"
	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func TestProcessor_Do(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		service := fixService("test", "test", "/test")
		client := new(automock.HttpClient)
		defer client.AssertExpectations(t)
		client.On("Do", mock.Anything).Return(fixHttpResponse(http.StatusOK, ""), nil).Once()

		processor := assethook.NewProcessor(2, client, false, testCallback(nil, nil), testCallback(nil, nil))

		// When
		result, err := processor.Do(context.TODO(), "./", []string{"processor_test.go"}, []v1alpha2.AssetWebhookService{service})

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(result).To(gomega.HaveLen(0))
	})

	t.Run("Error on service call", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		service := fixService("test", "test", "/test")
		client := new(automock.HttpClient)
		defer client.AssertExpectations(t)
		client.On("Do", mock.Anything).Return(fixHttpResponse(http.StatusInternalServerError, ""), nil).Once()

		processor := assethook.NewProcessor(2, client, false, testCallback(nil, nil), testCallback(nil, nil))

		// When
		_, err := processor.Do(context.TODO(), "./", []string{"processor_test.go"}, []v1alpha2.AssetWebhookService{service})

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
	})

	t.Run("Message on failed callback", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		service := fixService("test", "test", "/test")
		client := new(automock.HttpClient)
		defer client.AssertExpectations(t)
		client.On("Do", mock.Anything).Return(fixHttpResponse(http.StatusUnprocessableEntity, "message"), nil).Once()

		processor := assethook.NewProcessor(2, client, false, testCallback(nil, nil), testCallback([]string{"err"}, nil))

		// When
		result, err := processor.Do(context.TODO(), "./", []string{"processor_test.go"}, []v1alpha2.AssetWebhookService{service})

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(result).To(gomega.HaveLen(1))
	})

	t.Run("Error on failed callback", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		service := fixService("test", "test", "/test")
		client := new(automock.HttpClient)
		defer client.AssertExpectations(t)
		client.On("Do", mock.Anything).Return(fixHttpResponse(http.StatusUnprocessableEntity, "message"), nil).Once()

		processor := assethook.NewProcessor(2, client, false, testCallback(nil, nil), testCallback(nil, []error{fmt.Errorf("test")}))

		// When
		result, err := processor.Do(context.TODO(), "./", []string{"processor_test.go"}, []v1alpha2.AssetWebhookService{service})

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(result).To(gomega.HaveLen(0))
	})

	t.Run("Message on success callback", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		service := fixService("test", "test", "/test")
		client := new(automock.HttpClient)
		defer client.AssertExpectations(t)
		client.On("Do", mock.Anything).Return(fixHttpResponse(http.StatusOK, "message"), nil).Once()

		processor := assethook.NewProcessor(2, client, false, testCallback([]string{"err"}, nil), testCallback(nil, nil))

		// When
		result, err := processor.Do(context.TODO(), "./", []string{"processor_test.go"}, []v1alpha2.AssetWebhookService{service})

		// Then
		g.Expect(err).ToNot(gomega.HaveOccurred())
		g.Expect(result).To(gomega.HaveLen(1))
	})

	t.Run("Error on success callback", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		service := fixService("test", "test", "/test")
		client := new(automock.HttpClient)
		defer client.AssertExpectations(t)
		client.On("Do", mock.Anything).Return(fixHttpResponse(http.StatusOK, "message"), nil).Once()

		processor := assethook.NewProcessor(2, client, false, testCallback(nil, []error{fmt.Errorf("test")}), testCallback(nil, nil))

		// When
		result, err := processor.Do(context.TODO(), "./", []string{"processor_test.go"}, []v1alpha2.AssetWebhookService{service})

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(result).To(gomega.HaveLen(0))
	})

	t.Run("Invalid file", func(t *testing.T) {
		// Given
		g := gomega.NewGomegaWithT(t)

		service := fixService("test", "test", "/test")
		client := new(automock.HttpClient)
		defer client.AssertExpectations(t)

		processor := assethook.NewProcessor(2, client, false, testCallback(nil, nil), testCallback(nil, nil))

		// When
		result, err := processor.Do(context.TODO(), "./", []string{"xyz.go"}, []v1alpha2.AssetWebhookService{service})

		// Then
		g.Expect(err).To(gomega.HaveOccurred())
		g.Expect(result).To(gomega.HaveLen(0))
	})
}

func testCallback(messages []string, errors []error) assethook.Callback {
	return func(ctx context.Context, basePath, filePath string, responseBody io.Reader, messagesChan chan assethook.Message, errChan chan error) {
		for _, err := range errors {
			errChan <- err
		}

		for _, msg := range messages {
			messagesChan <- assethook.Message{Filename: filePath, Message: msg}
		}
	}
}

func fixService(name, namespace, endpoint string) v1alpha2.AssetWebhookService {
	return v1alpha2.AssetWebhookService{
		WebhookService: v1alpha2.WebhookService{
			Name:      name,
			Namespace: namespace,
			Endpoint:  endpoint,
		},
	}
}

func fixHttpResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       ioutil.NopCloser(strings.NewReader(body)),
	}
}
