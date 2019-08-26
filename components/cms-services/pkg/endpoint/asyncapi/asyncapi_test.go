package asyncapi_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/cms-services/pkg/endpoint/asyncapi"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/service/fake"
	"github.com/onsi/gomega"
)

func TestV1Validation(t *testing.T) {
	// Given
	g := gomega.NewWithT(t)

	srv, err := initService()
	g.Expect(err).ToNot(gomega.HaveOccurred())

	for testName, testCase := range map[string]struct {
		filePath       string
		expectedStatus int
	}{
		"valid yaml": {
			filePath:       "./v1/testdata/valid.yaml",
			expectedStatus: http.StatusOK,
		},
		"invalid yaml": {
			filePath:       "./v1/testdata/invalid.yaml",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		"valid json": {
			filePath:       "./v1/testdata/valid.json",
			expectedStatus: http.StatusOK,
		},
		"invalid json": {
			filePath:       "./v1/testdata/invalid.json",
			expectedStatus: http.StatusUnprocessableEntity,
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			// When
			response, err := serve(srv, testCase.filePath, "/v1/validate")

			// Then
			g.Expect(err).ToNot(gomega.HaveOccurred())
			g.Expect(response.StatusCode).To(gomega.Equal(testCase.expectedStatus))
		})
	}
}

func TestV1Conversion(t *testing.T) {
	// Given
	g := gomega.NewWithT(t)

	srv, err := initService()
	g.Expect(err).ToNot(gomega.HaveOccurred())

	for testName, testCase := range map[string]struct {
		filePath       string
		expectedStatus int
	}{
		"valid yaml": {
			filePath:       "./v1/testdata/anyof_1.1.0.yaml",
			expectedStatus: http.StatusOK,
		},
		"invalid yaml": {
			filePath:       "./v1/testdata/anyof_invalid.yaml",
			expectedStatus: http.StatusUnprocessableEntity,
		},
		"valid json": {
			filePath:       "./v1/testdata/anyof_1.1.0.json",
			expectedStatus: http.StatusOK,
		},
		"invalid json": {
			filePath:       "./v1/testdata/anyof_invalid.json",
			expectedStatus: http.StatusUnprocessableEntity,
		},
	} {
		t.Run(testName, func(t *testing.T) {
			g := gomega.NewGomegaWithT(t)
			// When
			response, err := serve(srv, testCase.filePath, "/v1/convert")

			// Then
			g.Expect(err).ToNot(gomega.HaveOccurred())
			g.Expect(response.StatusCode).To(gomega.Equal(testCase.expectedStatus))
		})
	}
}

func serve(srv *fake.Service, filePath string, endpoint string) (*http.Response, error) {
	body, contentType, err := fake.RequestBodyFromFile(filePath, "")
	if err != nil {
		return nil, err
	}
	response := srv.ServeHTTP(http.MethodPost, endpoint, contentType, body)
	return response, nil
}

func initService() (*fake.Service, error) {
	srv := fake.NewService()
	if err := asyncapi.AddToService(srv); err != nil {
		return nil, err
	}
	if err := srv.Start(context.TODO()); err != nil {
		return nil, err
	}
	return srv, nil
}
