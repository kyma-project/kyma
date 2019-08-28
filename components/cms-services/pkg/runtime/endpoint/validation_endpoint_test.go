package endpoint_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/endpoint"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/service/fake"
	"github.com/onsi/gomega"
)

func TestValidationEndpoint_Handle(t *testing.T) {
	for testName, testCase := range map[string]struct {
		targetMethod   string
		expectedStatus int
		filePath       string
		metadata       string
		validator      endpoint.Validator
	}{
		"OK": {
			expectedStatus: http.StatusOK,
			targetMethod:   http.MethodPost,
			filePath:       "./validation_endpoint.go",
			validator:      &fakeValidator{fail: false},
		},
		"invalid method": {
			expectedStatus: http.StatusMethodNotAllowed,
			targetMethod:   http.MethodGet,
		},
		"invalid request": {
			expectedStatus: http.StatusBadRequest,
			targetMethod:   http.MethodPost,
		},
		"missing file": {
			expectedStatus: http.StatusBadRequest,
			targetMethod:   http.MethodPost,
			metadata:       "{\"test\":\"\"}",
		},
		"validation failed": {
			expectedStatus: http.StatusUnprocessableEntity,
			targetMethod:   http.MethodPost,
			filePath:       "./validation_endpoint.go",
			validator:      &fakeValidator{fail: true},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)
			edp := endpoint.NewValidation("test", testCase.validator)
			body, contentType, err := fake.RequestBodyFromFile(testCase.filePath, testCase.metadata)
			g.Expect(err).ToNot(gomega.HaveOccurred())

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(edp.Handle)
			request := httptest.NewRequest(testCase.targetMethod, "/test", body)
			request.Header.Add("Content-Type", contentType)

			// when
			handler.ServeHTTP(recorder, request)

			// then
			g.Expect(recorder.Result().StatusCode).To(gomega.Equal(testCase.expectedStatus))
		})
	}
}

func TestValidationEndpoint_Handle_NoMultipartRequest(t *testing.T) {
	// given
	g := gomega.NewWithT(t)
	edp := endpoint.NewValidation("test", nil)

	recorder := httptest.NewRecorder()
	handler := http.HandlerFunc(edp.Handle)
	request := httptest.NewRequest(http.MethodPost, "/test", nil)

	// when
	handler.ServeHTTP(recorder, request)

	// then
	g.Expect(recorder.Result().StatusCode).To(gomega.Equal(http.StatusBadRequest))
}

var _ endpoint.Validator = &fakeValidator{}

type fakeValidator struct {
	fail bool
}

func (v *fakeValidator) Validate(ctx context.Context, reader io.Reader, metadata string) error {
	if v.fail {
		return errors.New("fail")
	}

	return nil
}
