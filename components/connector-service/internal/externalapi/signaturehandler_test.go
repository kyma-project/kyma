package externalapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/httperrors"

	"github.com/kyma-project/kyma/components/connector-service/internal/apperrors"
	"github.com/kyma-project/kyma/components/connector-service/internal/clientcontext"

	certMock "github.com/kyma-project/kyma/components/connector-service/internal/certificates/mocks"
	tokensMock "github.com/kyma-project/kyma/components/connector-service/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appName = "appName"
	token   = "token"

	host               = "host"
	country            = "country"
	organization       = "organization"
	organizationalUnit = "organizationalUnit"
	locality           = "locality"
	province           = "province"
)

var (
	tokenRequestRaw = compact([]byte("{\"csr\":\"Q1NSCg==\"}"))
	decodedCSR, _   = decodeStringFromBase64("Q1NSCg==")
	crtBase64       = "crtBase64"
)

func TestSignatureHandler_SignCSR(t *testing.T) {

	url := fmt.Sprintf("/v1/applications/certificates?token=%s", token)

	t.Run("should sign client certificate", func(t *testing.T) {
		// given

		tokenRemover := &tokensMock.Remover{}
		tokenRemover.On("Delete", token).Return()

		certService := &certMock.Service{}
		certService.On("SignCSR", decodedCSR, commonName).Return(crtBase64, nil)

		dummyClientContext := dummyClientContext{}
		connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return dummyClientContext, nil
		}

		signatureHandler := NewSignatureHandler(tokenRemover, certService, connectorClientExtractor, host)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		signatureHandler.SignCSR(rr, req)

		// then
		responseBody, err := ioutil.ReadAll(rr.Body)
		require.NoError(t, err)

		var certResponse certResponse
		err = json.Unmarshal(responseBody, &certResponse)
		require.NoError(t, err)

		assert.Equal(t, crtBase64, certResponse.CRT)
		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("should return 500 when failed to extract client context", func(t *testing.T) {
		// given
		errorExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return nil, apperrors.Internal("error")
		}

		signatureHandler := NewSignatureHandler(nil, nil, errorExtractor, host)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		signatureHandler.SignCSR(rr, req)

		// then
		errorResponse := readErrorResponse(t, rr.Body)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should return 500 when couldn't read request body", func(t *testing.T) {
		// given
		dummyClientContext := dummyClientContext{}
		connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return dummyClientContext, nil
		}

		signatureHandler := NewSignatureHandler(nil, nil, connectorClientExtractor, host)

		incorrectBody := []byte("incorrectBody")
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(incorrectBody))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		signatureHandler.SignCSR(rr, req)

		// then
		errorResponse := readErrorResponse(t, rr.Body)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("should return 400 when failed to decode base64", func(t *testing.T) {
		// given
		dummyClientContext := dummyClientContext{}
		connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return dummyClientContext, nil
		}

		signatureHandler := NewSignatureHandler(nil, nil, connectorClientExtractor, host)

		incorrectBase64Body := compact([]byte("{\"csr\":\"not base 64\"}"))
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(incorrectBase64Body))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		signatureHandler.SignCSR(rr, req)

		// then
		errorResponse := readErrorResponse(t, rr.Body)

		assert.Equal(t, http.StatusBadRequest, errorResponse.Code)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("should return 500 when failed to sign CSR", func(t *testing.T) {
		// given
		certService := &certMock.Service{}
		certService.On("SignCSR", decodedCSR, commonName).Return("", apperrors.Internal("error"))

		dummyClientContext := dummyClientContext{}
		connectorClientExtractor := func(ctx context.Context) (clientcontext.ConnectorClientContext, apperrors.AppError) {
			return dummyClientContext, nil
		}

		signatureHandler := NewSignatureHandler(nil, certService, connectorClientExtractor, host)

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(tokenRequestRaw))
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		// when
		signatureHandler.SignCSR(rr, req)

		// then
		errorResponse := readErrorResponse(t, rr.Body)

		assert.Equal(t, http.StatusInternalServerError, errorResponse.Code)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func readErrorResponse(t *testing.T, body io.Reader) httperrors.ErrorResponse {
	responseBody, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var errorResponse httperrors.ErrorResponse
	err = json.Unmarshal(responseBody, &errorResponse)
	require.NoError(t, err)

	return errorResponse
}

func compact(src []byte) []byte {
	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, src)
	if err != nil {
		return src
	}
	return buffer.Bytes()
}
