package internalapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-project/kyma/components/connector-service/internal/revocation/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRevocationHandler(t *testing.T) {

	urlRevocation := "/v1/applications/certificates/revocations"
	hashedTestCert := "f21139ef2b82d02ee73a56c5c73c053fbafa3480a0b35459cba276b0667c57fc"

	t.Run("should revoke certificate and return http code 201", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(nil)

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		revocationBody := revocationBody{
			Hash: hashedTestCert,
		}

		body, err := marshall(revocationBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, urlRevocation, body)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})

	t.Run("should return http code 201 when certificate already revoked", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(nil)

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		revocationBody := revocationBody{
			Hash: hashedTestCert,
		}
		body, err := marshall(revocationBody)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, urlRevocation, body)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)

		//when
		body, err = marshall(revocationBody)
		require.NoError(t, err)
		req = httptest.NewRequest(http.MethodPost, urlRevocation, body)

		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusCreated, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})

	t.Run("should return http code 400 when certificate hash not passed", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		req := httptest.NewRequest(http.MethodPost, urlRevocation, nil)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		revocationListRepository.AssertNotCalled(t, "Insert", mock.AnythingOfType("string"))
	})

	t.Run("should return http code 400 when failed to unmarshall", func(t *testing.T) {
		//given

		revocationListRepository := &mocks.RevocationListRepository{}

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		testPayload := struct {
			Key string
		}{
			Key: "key",
		}

		body, err := marshall(testPayload)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, urlRevocation, body)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		revocationListRepository.AssertNotCalled(t, "Insert", mock.AnythingOfType("string"))
	})

	t.Run("should return http code 500 when certificate revocation not persisted", func(t *testing.T) {
		//given
		revocationListRepository := &mocks.RevocationListRepository{}
		revocationListRepository.On("Insert", hashedTestCert).Return(errors.New("Error"))

		handler := NewRevocationHandler(revocationListRepository)

		rr := httptest.NewRecorder()

		revocationBody := revocationBody{
			Hash: hashedTestCert,
		}
		body, err := marshall(revocationBody)
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPost, urlRevocation, body)

		//when
		handler.Revoke(rr, req)

		//then
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		revocationListRepository.AssertExpectations(t)
	})
}

func marshall(body interface{}) (io.Reader, error) {
	var b bytes.Buffer

	err := json.NewEncoder(&b).Encode(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b.Bytes()), nil
}
