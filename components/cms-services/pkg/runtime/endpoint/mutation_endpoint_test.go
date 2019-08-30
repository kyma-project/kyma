package endpoint_test

import (
	"bytes"
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

func TestMutationEndpoint_Handle(t *testing.T) {
	message := "test response"

	for testName, testCase := range map[string]struct {
		targetMethod   string
		expectedStatus int
		filePath       string
		metadata       string
		mutator        endpoint.Mutator
		response       string
	}{
		"OK": {
			expectedStatus: http.StatusOK,
			targetMethod:   http.MethodPost,
			filePath:       "./mutation_endpoint.go",
			mutator:        &fakeMutator{fail: false, message: message},
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
			mutator:        &fakeMutator{fail: true},
		},
		"no changes": {
			expectedStatus: http.StatusNotModified,
			targetMethod:   http.MethodPost,
			filePath:       "./validation_endpoint.go",
			mutator:        &fakeMutator{noChanges: true},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			// given
			g := gomega.NewWithT(t)
			edp := endpoint.NewMutation("test", testCase.mutator)
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
			if recorder.Result().StatusCode == http.StatusOK {
				buffer := new(bytes.Buffer)
				buffer.ReadFrom(recorder.Result().Body)

				g.Expect(buffer.String()).To(gomega.Equal(message))
			}
		})
	}
}

func TestMutationEndpoint_Handle_NoMultipartRequest(t *testing.T) {
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

var _ endpoint.Mutator = &fakeMutator{}

type fakeMutator struct {
	fail      bool
	noChanges bool
	message   string
}

func (m *fakeMutator) Mutate(ctx context.Context, reader io.Reader, metadata string) ([]byte, bool, error) {
	if m.fail {
		return nil, false, errors.New("fail")
	}

	if m.noChanges {
		return nil, false, nil
	}

	return []byte(m.message), true, nil
}
